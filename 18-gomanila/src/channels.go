package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	c := make(chan int)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(z int) {
			defer wg.Done()
			time.Sleep(1 * time.Second)
			c <- z
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case z := <-c:
				fmt.Printf(">> z: %d\n", z)
			case <-time.After(5 * time.Second):
				return
			}
		}
	}()

	wg.Wait()
	fmt.Println("done")
}
