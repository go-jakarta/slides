package main

import (
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
)

// START OMIT
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
    user: [User!]!
  }
`

// END OMIT

func main() {
	schema := graphql.MustParseSchema(schemaString, nil).Inspect()
	for _, typ := range schema.Types() {
		fmt.Printf("type: %s (%s)\n", *typ.Name(), typ.Kind())
	}
	typ := schema.QueryType()
	fmt.Println("query:")
	for _, field := range *typ.Fields(nil) {
		fmt.Printf(" field: %s\n", field.Name())
	}
}
