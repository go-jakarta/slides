package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brankas/goji"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/relay"
	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// START1 OMIT
const schemaString string = `
  scalar Time
  type User {
    id: ID!
    email: String!
    name: String
    dob: Time
    orgs: [Org!]!
  }
  type Org {
    name: String!
  }
  type Query {
	users(name: String!): [User!]!
    orgs(name: String!): [Org!]!
  }
`

// END1 OMIT

type User struct {
	ID    graphql.ID
	Email string
	Name  *string
	Dob   *graphql.Time
	Orgs  []*Org
}

type Org struct {
	Name string
}

func run() error {
	if err := os.RemoveAll("test.db"); err != nil && !os.IsNotExist(err) {
		return err
	}
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		return err
	}
	defer db.Close()
	s := &server{Mux: goji.New(), db: db}
	s.schema, err = graphql.ParseSchema(schemaString, s, graphql.UseFieldResolvers())
	if err != nil {
		return err
	}
	if err := s.init(); err != nil {
		return err
	}
	s.Handle(goji.NewPathSpec("/query"), &relay.Handler{Schema: s.schema})
	return http.ListenAndServe(":8080", s)
}

type server struct {
	*goji.Mux
	db     *sql.DB
	schema *graphql.Schema
}

func (s *server) init() error {
	// START2 OMIT
	for _, typ := range s.schema.Inspect().Types() {
		name := *typ.Name()
		if typ.Kind() != "OBJECT" || strings.HasPrefix(name, "__") ||
			strings.HasSuffix(name, "Query") {
			continue
		}
		ss := fmt.Sprintf("create table %s (", name)
		for i, field := range *typ.Fields(nil) {
			if i != 0 {
				ss += ","
			}
			ss += fmt.Sprintf("\n  %s %s", field.Name(), sqlType(field))
		}
		ss += "\n)"
		log.Printf("adding %s (%s):\n%s", *typ.Name(), typ.Kind(), ss)
		if _, err := s.db.Exec(ss); err != nil {
			return err
		}
	}
	// END2 OMIT

	// insert some data
	if _, err := s.db.Exec(`
	  insert into org (name) values ('org1'), ('org2')
`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
	  insert into user (name) values ('user1'), ('user2')
`); err != nil {
		return err
	}

	return nil
}

// START3 OMIT
func (s *server) Orgs(args struct{ Name string }) ([]*Org, error) {
	sqlstr := `select * from org`
	var sqlargs []interface{}
	if args.Name != "" {
		sqlstr, sqlargs = sqlstr+` where name = ?`, append(sqlargs, args.Name)
	}
	log.Printf("executing:\n%s\n%v", sqlstr, sqlargs)
	rows, err := s.db.Query(sqlstr, sqlargs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []*Org
	for rows.Next() {
		org := new(Org)
		if err := rows.Scan(&org.Name); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}

// END3 OMIT

func (s *server) Users(args struct{ Name string }) ([]*User, error) {
	return nil, nil
}

func sqlType(field *introspection.Field) string {
	return "text"
}
