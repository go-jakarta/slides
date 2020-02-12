package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	return nil
}
