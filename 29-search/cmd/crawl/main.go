package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bgadrian/data-structures/priorityqueue"
	prq "github.com/bgadrian/data-structures/priorityqueue"
	"github.com/kenshaw/diskcache"
	_ "github.com/lib/pq"
	"github.com/mitchellh/go-homedir"
	"github.com/xo/dburl"
	"github.com/zellyn/kooky"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/sync/errgroup"
)

const imdbBase = "https://www.thetvdb.com"

func main() {
	flagDB := flag.String("db", "", "database url")
	flagWorkers := flag.Int("workers", runtime.NumCPU(), "workers")
	flag.Parse()
	if err := run(*flagDB, *flagWorkers); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

type Task interface {
	Exec(ctx context.Context, id int) error
}

type state struct {
	db            *sql.DB
	cl            *http.Client
	errCount      int32
	totalErrCount int32
	tasks         *prq.HierarchicalQueue
}

// run is the high level crawl implementation.
func run(dsn string, workers int) error {
	// ensure flags have been set
	if dsn == "" {
		dsn = os.Getenv("DB")
	}
	if dsn == "" {
		return errors.New("must provide -db or $ENV{DB}")
	}
	db, err := dburl.Open(dsn)
	if err != nil {
		return err
	}

	// setup state
	st := &state{
		db:    db,
		tasks: priorityqueue.NewHierarchicalQueue(20, false),
	}
	st.cl, err = buildClient()
	if err != nil {
		return err
	}
	st.tasks.Enqueue(newScrape(st, imdbBase), 0)

	ctx, cancel := context.WithCancel(context.Background())

	// start workers
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < workers; i++ {
		eg.Go(st.worker(ctx, i))
	}

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- eg.Wait()
	}()

	// error checker
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	var lastErr error
loop:
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && err != context.Canceled {
				log.Printf("context canceled: %v", err)
			}
			cancel()
			break loop

		case v := <-sig:
			log.Printf("got signal: %v", v)
			cancel()
			break loop

		case <-t.C:
			errCount := atomic.SwapInt32(&st.errCount, 0)
			totalErrCount := atomic.AddInt32(&st.totalErrCount, errCount)
			switch {
			case errCount >= 20:
				cancel()
				lastErr = errors.New("errors over last tick exceeds 20")
				break loop
			case totalErrCount >= 500:
				cancel()
				lastErr = fmt.Errorf("exceeded 500 total errors (%d)", totalErrCount)
				break loop
			}
		}
	}

	if lastErr != nil {
		log.Printf("ERROR: %v", lastErr)
	}

	log.Printf("END: waiting for workers to terminate")
	if lastErr := <-done; lastErr != nil {
		log.Printf("ERROR: END: worker errgroup: %v", lastErr)
	}

	log.Printf("END: total errors: %d", atomic.LoadInt32(&st.totalErrCount))
	return nil
}

// buildClient builds a HTTP client for use with crawling
func buildClient() (*http.Client, error) {
	cacheDir, err := homedir.Expand("~/imdbcache")
	d, err := diskcache.New(
		diskcache.WithBasePathFs(cacheDir),
		diskcache.WithErrorTruncator(),
		diskcache.WithMinifier(),
		diskcache.WithTTL(24*time.Hour),
		diskcache.WithGzipCompression(),
		diskcache.WithLongPathHandler(func(s string) string {
			return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
		}),
	)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(imdbBase)
	if err != nil {
		return nil, err
	}
	cookiePath, err := homedir.Expand("~/.config/google-chrome/Default/Cookies")
	if err != nil {
		return nil, err
	}
	cookies, err := kooky.ReadChromeCookies(cookiePath, "imdb.com", "", time.Time{})
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	var l []*http.Cookie
	for _, c := range cookies {
		v := c.HttpCookie()
		l = append(l, &v)
	}
	jar.SetCookies(u, l)
	return &http.Client{
		Transport: d,
		Jar:       jar,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func (st *state) worker(ctx context.Context, id int) func() error {
	return func() error {
		log.Printf("WORKER %d: starting", id)

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop

			default:
				v, err := st.tasks.Dequeue()
				if err != nil && err.Error() != "the queue is empty" {
					log.Printf("WORKER %d: DEQUEUE ERROR: %v", id, err)
				}
				if v != nil {
					if err := v.(Task).Exec(ctx, id); err != nil {
						log.Printf("WORKER %d: TASK EXEC ERROR: %v", id, err)
						atomic.AddInt32(&st.errCount, 1)
					}
				}
			}
		}

		log.Printf("WORKER %d: done", id)
		return nil
	}
}

type scrape struct {
	st     *state
	urlstr string
}

func newScrape(st *state, urlstr string) *scrape {
	return &scrape{
		st:     st,
		urlstr: urlstr,
	}
}

func (t *scrape) Exec(ctx context.Context, id int) error {
	if !strings.HasPrefix(t.urlstr, imdbBase) {
		log.Printf("%d: SKIPPING: %s", id, t.urlstr)
	}

	log.Printf("RETRIEVING: %s", t.urlstr)
	req, err := http.NewRequest("GET", t.urlstr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36")
	res, err := t.st.cl.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	boundary := bytes.Index(buf, crlfcrlf)
	if boundary == -1 {
		return fmt.Errorf("%d: could not find boundary in response for %s", id, t.urlstr)
	}
	return t.st.tasks.Enqueue(newProcess(t.st, t.urlstr, buf[:boundary], buf[boundary+4:]), 10)
}

var urlRE = regexp.MustCompile(`/+`)
var spaceRE = regexp.MustCompile(`\s+`)
var crlfcrlf = []byte("\r\n\r\n")

type process struct {
	st     *state
	urlstr string
	header []byte
	body   []byte
}

func newProcess(st *state, urlstr string, header, body []byte) *process {
	return &process{
		st:     st,
		urlstr: urlstr,
		header: header,
		body:   body,
	}
}

func (t *process) Exec(ctx context.Context, id int) error {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(t.body))
	if err != nil {
		return err
	}
	title := doc.Find("title").Text()
	words := doc.Find("body").Text()

	var hrefs []string
	doc.Find(`a`).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		hrefLower := strings.ToLower(href)
		switch {
		case strings.HasPrefix(hrefLower, "https://") || strings.HasPrefix(hrefLower, "http://"):
		case strings.HasPrefix(hrefLower, "/"):
			href = imdbBase + urlRE.ReplaceAllString(href, "/")
		default:
			href = t.urlstr + urlRE.ReplaceAllString(href, "/")
		}
		if strings.HasPrefix(href, imdbBase+"/") {
			hrefs = append(hrefs, href)
		}
	})

	// clean up title + words
	title = strings.TrimSpace(spaceRE.ReplaceAllString(title, " "))
	words = strings.TrimSpace(spaceRE.ReplaceAllString(words, " "))

	if err := t.st.tasks.Enqueue(newInsert(t.st, t.urlstr, t.header, t.body, title, words), 1); err != nil {
		return err
	}
	for _, href := range hrefs {
		if err := t.st.tasks.Enqueue(newScrape(t.st, href), 10); err != nil {
			return err
		}
	}
	return nil
}

type insert struct {
	st     *state
	urlstr string
	header []byte
	body   []byte
	title  string
	words  string
}

func newInsert(st *state, urlstr string, header, body []byte, title, words string) *insert {
	return &insert{
		st:     st,
		urlstr: urlstr,
		header: header,
		body:   body,
		title:  title,
		words:  words,
	}
}

func (t *insert) Exec(ctx context.Context, id int) error {
	// add to db
	const sqlstr = `INSERT INTO pages (` +
		`location,header,body,title,words` +
		`) VALUES (` +
		`$1,$2,$3,$4,$5` +
		`) ` +
		`ON CONFLICT(location) DO ` +
		`UPDATE SET ` +
		`header = $2, body = $3, title = $4, words = $5, refreshed = NOW() ` +
		`WHERE EXCLUDED.location = $1`

	_, err := t.st.db.ExecContext(
		ctx, sqlstr, t.urlstr, t.header, t.body, t.title, t.words,
	)
	return err
}
