package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "hello secure world")
	})

	log.Fatal(http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil))
}
