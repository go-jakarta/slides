package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		name := "[unknown]"
		if n := req.URL.Query().Get("name"); n != "" {
			name = n
		}
		fmt.Fprintf(res, "hello %s!", name)
	})
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
