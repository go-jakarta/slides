package main

import (
	"log"
	"time"
)

func main() {
	ch := make(chan int)
	go func() {
		for i := range ch {
			log.Printf("got: %d", i)
		}
	}()

	for i := 0; i < 15; i++ {
		ch <- i
	}
	close(ch)
	time.Sleep(1 * time.Second)
}
