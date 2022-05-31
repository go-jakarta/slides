package main

import "fmt"

type Number interface {
	int8 | uint8 | uint | uint32 | float32 | float64
}

type MySlice[T Number] []T

func MyPrint[N Number, T MySlice[N]](v T) {
	var n N
	for _, i := range v {
		n += i
	}
	fmt.Printf("%T: %v\n", v, n)
}

func main() {
	MyPrint(MySlice[float64]{1.0, 2.0, 3.1})
	MyPrint(MySlice[uint]{1, 2, 3})
}
