package main

//go:generate ./models/gen.sh

import (
	"flag"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/xo/dburl"

	"github.com/go-jakarta/slides/27-aws/webapp/services"
)

var (
	flagListen = flag.String("l", "0.0.0.0:8080", "listen")
	flagDB     = flag.String("db", "pg://postgres:P4ssw0rd@localhost?sslmode=disable", "database")
)

func main() {
	flag.Parse()
	if err := run(*flagListen, *flagDB); err != nil {
		log.Fatal("error: %v", err)
	}
}

func run(addr string, dsn string) error {
	log.Printf("listening: http://%s database: %s", addr, dsn)
	db, err := dburl.Open(dsn)
	if err != nil {
		return err
	}
	s := services.NewServer(db, log.Printf)
	return http.ListenAndServe(addr, s)
}
