package main

import (
	"fmt"
	"net/http"
)

type counter int64

func (c *counter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	*c++
	fmt.Fprintf(res, "total requests: %d", *c)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "hello multiplexed world!")
	})
	mux.Handle("/counter", new(counter))
	http.ListenAndServe(":8000", mux)
}
