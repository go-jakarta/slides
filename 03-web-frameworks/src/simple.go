package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "hello world")
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}
