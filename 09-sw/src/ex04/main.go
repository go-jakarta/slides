package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		if name := req.URL.Query().Get("hi"); name != "" {
			fmt.Fprintf(res, "hi %s!", name)
		} else {
			fmt.Fprint(res, "no one to say hi to")
		}
	})

	log.Fatal(http.Serve(autocert.NewListener("example.com"), mux))
}
