package main

import "fmt"

type (
	tF  func(int) int
	tRF func(tF) tF
	tX  func(tX) tF
)

func Y(rf tRF) tF {
	return func(x tX) tF { return func(n int) int { return rf(x(x))(n) } }(
		func(x tX) tF { return func(n int) int { return rf(x(x))(n) } })
}

var fib = Y(func(f tF) tF {
	return func(n int) int {
		if n == 0 || n == 1 {
			return n
		}
		return f(n-1) + f(n-2)
	}
})

var fact = Y(func(f tF) tF {
	return func(n int) int {
		if n < 2 {
			return 1
		}
		return n * f(n-1)
	}
})

func main() {
	for i := 1; i < 10; i++ {
		fmt.Printf("%2d : %2d, %6d\n", i, fib(i), fact(i))
	}
}
