package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	goji "goji.io"
	"goji.io/pat"

	"gophers.id/slides/22-aks/webapp/models"
)

type Server struct {
	db   *sql.DB
	logf func(string, ...interface{})
}

func NewServer(db *sql.DB, f func(string, ...interface{})) *Server {
	return &Server{db: db, logf: f}
}

func (s *Server) Handler() *goji.Mux {
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/authors"), s.GetAuthorsByName)

	mux.HandleFunc(pat.Post("/author"), s.CreateAuthor)
	mux.HandleFunc(pat.Get("/author/:id"), s.GetAuthorByID)
	mux.HandleFunc(pat.Put("/author/:id"), s.UpdateAuthor)
	mux.HandleFunc(pat.Delete("/author/:id"), s.DeleteAuthor)

	return mux
}

func (s *Server) GetAuthorsByName(res http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if name == "" {
		jsonerrorf(res, http.StatusBadRequest, "missing name")
		return
	}
	authors, err := models.AuthorsByName(s.db, name)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusNotFound, "not found")
		return
	}
	jsonf(res, map[string]interface{}{
		"authors": authors,
	})
}

func (s *Server) GetAuthorByID(res http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusBadRequest, err.Error())
		return
	}
	author, err := models.AuthorByAuthorID(s.db, id)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusNotFound, "not found")
		return
	}
	jsonf(res, author)
}

func (s *Server) CreateAuthor(res http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	author := new(models.Author)
	if err := dec.Decode(author); err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusBadRequest, "invalid author")
		return
	}
	if err := author.Insert(s.db); err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusInternalServerError, "unable to create author")
		return
	}
	jsonf(res, map[string]interface{}{
		"author_id": author.AuthorID,
	})
}

func (s *Server) UpdateAuthor(res http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	id, err := getID(req)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusBadRequest, err.Error())
		return
	}
	author, err := models.AuthorByAuthorID(s.db, id)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusNotFound, "author not found")
		return
	}
	if err := dec.Decode(author); err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusBadRequest, "invalid author")
		return
	}
	author.AuthorID = id
	if err := author.Update(s.db); err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusInternalServerError, "unable to update author")
		return
	}
	jsonf(res, map[string]interface{}{
		"author_id": author.AuthorID,
	})
}

func (s *Server) DeleteAuthor(res http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusBadRequest, err.Error())
		return
	}
	author, err := models.AuthorByAuthorID(s.db, id)
	if err != nil {
		s.logf("error: %v", err)
		jsonerrorf(res, http.StatusNotFound, "author not found")
		return
	}
	if err := author.Delete(s.db); err != nil {
		jsonerrorf(res, http.StatusInternalServerError, "unable to delete author")
		return
	}
}

func jsonerrorf(res http.ResponseWriter, code int, s string, v ...interface{}) {
	buf, err := json.Marshal(map[string]string{
		"error": fmt.Sprintf(s, v...),
	})
	if err != nil {
		http.Error(res, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}
	http.Error(res, string(buf), code)
}

func jsonf(res http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(res).Encode(v); err != nil {
		http.Error(res, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}
}

func getID(req *http.Request) (int, error) {
	idstr := pat.Param(req, "id")
	if idstr == "" {
		return 0, errors.New("missing id")
	}
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
