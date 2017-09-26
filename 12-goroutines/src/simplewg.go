package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			time.Sleep(time.Second * time.Duration(rand.Int()%3))
			fmt.Printf("goroutine: %d\n", i)
		}(wg, i)
	}
	wg.Wait()
}
