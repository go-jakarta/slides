package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"search/assets"
	"search/assets/templates"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/brankas/goji"
	"github.com/brankas/sentinel"
	"github.com/gorilla/csrf"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

func main() {
	flagDB := flag.String("db", "", "database url")
	flagListen := flag.String("listen", "", "listen address")
	flag.Parse()
	if err := run(*flagDB, *flagListen); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run is the main application entry point.
func run(dsn, listen string) error {
	// ensure flags have been set
	if dsn == "" {
		dsn = os.Getenv("DB")
	}
	if dsn == "" {
		return errors.New("must provide -db or $ENV{DB}")
	}
	if listen == "" {
		listen = os.Getenv("LISTEN")
	}
	if listen == "" {
		listen = ":3000"
	}

	// set template stuff
	templates.Asset = func(fn string) string {
		return "/_/" + assets.ManifestPath("/"+strings.TrimPrefix(fn, "/"))
	}
	templates.CsrfToken = func(req *http.Request) string {
		return string(csrf.TemplateField(req))
	}

	// open database
	db, err := dburl.Open(dsn)
	if err != nil {
		return err
	}

	// listen
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}

	// run server
	s, err := sentinel.New(
		sentinel.Logf(log.Printf),
	)
	if err := s.HTTP(l, newServer(db)); err != nil {
		return err
	}
	return s.Run(context.Background())
}

// Server is a search server.
type Server struct {
	db *sql.DB
	*goji.Mux
}

// newServer creates a new search server.
func newServer(db *sql.DB) *Server {
	s := &Server{
		db:  db,
		Mux: goji.New(),
	}
	s.Use(gziphandler.GzipHandler)
	s.Handle(goji.NewPathSpec("/_/*"), assets.StaticHandler(goji.Path))

	s.HandleFunc(goji.Get("/find"), s.find)
	s.HandleFunc(goji.Get("/"), s.index)
	return s
}

// index serves the index page.
func (s *Server) index(res http.ResponseWriter, req *http.Request) {
	templates.Do(res, req, &templates.IndexPage{})
}

// find handles retrieving results from the database and serving to end user.
func (s *Server) find(res http.ResponseWriter, req *http.Request) {
	results, err := s.search(req.Context(), req.URL.Query().Get("q"))
	if err != nil {
		log.Printf("ERROR: %v", err)
		http.Error(res, "unable to execute query", http.StatusInternalServerError)
		return
	}
	templates.Do(res, req, &templates.FindPage{Results: results})
}

// search executes a fulltext search against the database.
func (s *Server) search(ctx context.Context, query string) ([]templates.Result, error) {
	const sqlstr = `SELECT ` +
		`page_id, ` +
		`location, ` +
		`title, ` +
		`ts_headline('english', words, qq) AS summary, ` +
		`ts_rank_cd(words_tsv, qq) AS rank ` +
		`FROM pages, websearch_to_tsquery('english', $1) AS qq ` +
		`WHERE words_tsv @@ qq ` +
		`ORDER BY rank DESC ` +
		`LIMIT 10`

	// query
	res, err := s.db.QueryContext(ctx, sqlstr, query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// load results
	var results []templates.Result
	for res.Next() {
		if err := res.Err(); err != nil {
			return nil, err
		}
		var r templates.Result
		if err := res.Scan(&r.PageID, &r.Location, &r.Title, &r.Summary, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
