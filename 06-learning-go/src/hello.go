package main

import "fmt"

// program entry point
func main() {
	str := "hello 世界"
	for _, s := range str {
		fmt.Printf("%c", s)
	}
	fmt.Println()
}
