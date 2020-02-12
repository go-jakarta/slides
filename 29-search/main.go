package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run is the main application entry point.
func run() error {
}

type Server struct {
}
