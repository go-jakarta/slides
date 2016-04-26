package main

import (
	"fmt"
	"net/http"
)

// START OMIT
type counter int64

func (c *counter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	*c++
	fmt.Fprintf(res, "total requests: %d", *c)
}

func main() {
	http.ListenAndServe(":8000", new(counter))
}

// END OMIT
