package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-jakarta/slides/03-web-frameworks/src/models"
	_ "github.com/mattn/go-sqlite3"
)

// LOGGER OMIT
type logger struct {
	logger io.Writer
	next   http.Handler
}

func (l *logger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(l.logger, "got request from %s for %s\n", req.RemoteAddr, req.URL.Path)
	l.next.ServeHTTP(res, req)
}

// END OMIT

var db models.XODB

func main() {
	var err error

	models.XOLog = func(s string, p ...interface{}) {
		fmt.Printf("> SQL: %s -- %v\n", s, p)
	}

	// open database
	db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}

	// MUX OMIT
	// create serve mux
	mux := http.NewServeMux()
	mux.HandleFunc("/authors", listAuthors)
	mux.HandleFunc("/add", addAuthor)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/", index)

	// serve
	log.Fatal(http.ListenAndServe(":8000", &logger{
		logger: os.Stderr,
		next:   mux,
	}))
	// END OMIT
}

func writeJSON(res http.ResponseWriter, data interface{}, code int) {
	buf, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(code)
	fmt.Fprintf(res, string(buf))
}

func listAuthors(res http.ResponseWriter, req *http.Request) {
	// get authors
	authors, err := models.GetAuthors(db)
	if err != nil {
		writeJSON(res, map[string]interface{}{
			"error": "could not read database",
		}, http.StatusInternalServerError)
		return
	}

	// output
	writeJSON(res, authors, http.StatusOK)
}

func addAuthor(res http.ResponseWriter, req *http.Request) {
	// process data
	name := req.URL.Query().Get("name")
	author := &models.Author{
		Name: name,
	}

	// save to database
	err := author.Save(db)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// write result
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "OK")
}

var tpl = template.Must(template.New("").Parse(`hello {{.}}!`))

func index(res http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if name == "" {
		name = "[unknown]"
	}
	tpl.Execute(res, name)
}
