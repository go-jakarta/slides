package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("hello world: %s %s\n", runtime.GOOS, runtime.GOARCH)
}
