package templates

import (
	"net/http"
	"time"

	qtpl "github.com/valyala/quicktemplate"
)

type AssetFunc func(path string) string
type CsrfTokenFunc func(req *http.Request) string

var (
	Asset     AssetFunc
	CsrfToken CsrfTokenFunc
	year      string
	jakarta   *time.Location
)

func init() {
	time.Local = time.UTC

	var err error
	jakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}

	year = time.Now().In(jakarta).Format("2006")

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			year = time.Now().In(jakarta).Format("2006")
		}
	}()
}

// Do displays the page using the supplied translation func and translations.
func Do(res http.ResponseWriter, req *http.Request, p Page) {
	w := qtpl.AcquireWriter(res)
	defer qtpl.ReleaseWriter(w)
	StreamLayoutPage(w, p, req)
}

// Result is a search result.
type Result struct {
	PageID   string
	Location string
	Title    string
	Summary  string
	Rank     float64
}
