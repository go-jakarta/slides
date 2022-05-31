package main

import "fmt"

type Number interface {
	~int | ~int8 | ~uint8 | ~uint | ~uint32 | ~float32 | ~float64
}

type MySlice[N Number] []N

func MyPrint[N Number](v ...N) {
	var n N
	for _, i := range v {
		n += i
	}
	fmt.Printf("%T: %v\n", v, n)
}

func main() {
	MyPrint(MySlice[float64]{1.0, 2.0, 3.1}...)
	MyPrint(1, 2, 3)
	v := []float32{1.1, 2.2, 3.3}
	MyPrint(v...)
}
