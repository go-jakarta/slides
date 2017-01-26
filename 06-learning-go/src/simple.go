package main

import (
	"fmt"
	"os"
)

func main() {
	for i, a := range os.Args {
		fmt.Printf(">> arg %d: %s\n", i, a)
	}
}
