package main

//go:generate ./models/gen.sh

import (
	"flag"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/xo/dburl"
	"gophers.id/slides/22-aks/webapp/services"
)

var (
	flagListen = flag.String("l", "0.0.0.0:8080", "listen")
	flagDB     = flag.String("db", "pg://postgres:P4ssw0rd@localhost", "database")
)

func main() {
	flag.Parse()

	log.Printf("database: %s", *flagDB)
	if err := run(); err != nil {
		log.Fatal("error: %v", err)
	}
}

func run() error {
	db, err := dburl.Open(*flagDB)
	if err != nil {
		return err
	}
	s := services.NewServer(db, log.Printf)
	return http.ListenAndServe(*flagListen, s.Handler())
}
