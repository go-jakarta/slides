package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// LOGGER OMIT
type logger struct {
	logger io.Writer
	next   http.Handler
}

func (l *logger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(l.logger, "got request from %s for %s\n", req.RemoteAddr, req.URL.Path)
	l.next.ServeHTTP(res, req)
	fmt.Fprintf(l.logger, "returning counter: %d\n", *c)
}

// END OMIT

type counter int64

func (c *counter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	*c++
	fmt.Fprintf(res, "total requests: %d", *c)
}

// MAIN OMIT
var c = new(counter)

func main() {
	mw := &logger{
		logger: os.Stderr,
		next:   c,
	}

	http.ListenAndServe(":8000", mw)
}

// END OMIT
