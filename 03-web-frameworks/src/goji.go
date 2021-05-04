package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-jakarta/slides/03-web-frameworks/src/models"
	_ "github.com/mattn/go-sqlite3"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

var db models.XODB

func init() {
	var err error

	models.XOLog = func(s string, p ...interface{}) {
		fmt.Printf("> SQL: %s -- %v\n", s, p)
	}

	// open database
	db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
}

func writeJSON(res http.ResponseWriter, data interface{}, code int) {
	buf, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(code)
	fmt.Fprintf(res, string(buf))
}

func main() {
	mux := goji.NewMux()

	mux.HandleFuncC(pat.Get("/hello/:name"), func(c context.Context, res http.ResponseWriter, req *http.Request) {
		name := pat.Param(c, "name")
		fmt.Fprintf(res, "hello, %s!", name)
	})

	mux.HandleFuncC(pat.Get("/authors"), func(c context.Context, res http.ResponseWriter, req *http.Request) {
		// get authors
		authors, err := models.GetAuthors(db)
		if err != nil {
			writeJSON(res, map[string]interface{}{
				"error": err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		writeJSON(res, map[string]interface{}{
			"authors": authors,
		}, http.StatusOK)
	})

	// SPLIT OMIT

	mux.HandleFuncC(pat.Post("/add"), func(c context.Context, res http.ResponseWriter, req *http.Request) {
		name := req.PostFormValue("name")
		author := &models.Author{
			Name: name,
		}

		// save to database
		err := author.Save(db)
		if err != nil {
			writeJSON(res, map[string]interface{}{
				"error": err.Error(),
			}, http.StatusInternalServerError)
		}

		writeJSON(res, map[string]interface{}{
			"id": author.AuthorID,
		}, http.StatusOK)
	})

	http.ListenAndServe(":8000", mux)
}
