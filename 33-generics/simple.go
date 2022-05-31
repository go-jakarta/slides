package main

import "fmt"

func sum[T int8 | int | int32 | int64 | float32 | float64](v ...T) T {
	var a T
	for _, i := range v {
		a += i
	}
	return a
}

func main() {
	fmt.Printf("float64: %f\n", sum(1.0, 2.0, 3.1))
	fmt.Printf("int: %d\n", sum(1, 2, 3))
	// compile error
	fmt.Printf("uint: %d\n", sum(uint32(1), uint32(2), uint32(3)))
}
