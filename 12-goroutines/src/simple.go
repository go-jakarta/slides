package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(time.Second * time.Duration(rand.Int()%3))
			fmt.Printf("goroutine: %d\n", i)
		}(i)
	}
	time.Sleep(50 * time.Second)
}
