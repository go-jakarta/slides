package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	flagListen = flag.String("l", ":8080", "listen")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "hello!\n")
	})

	log.Fatal(http.ListenAndServe(*flagListen, mux))
}
