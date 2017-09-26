package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

func worker(wg *sync.WaitGroup, i int, ch chan string) {
	log.Printf("worker %d started", i)
	for s := range ch {
		log.Printf("worker %d processing %s", i, s)
		wg.Done()
	}
}

func main() {
	wg := new(sync.WaitGroup)
	ch := make(chan string)
	for i := 0; i < 4; i++ {
		go worker(wg, i, ch)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 16; i++ {
			s := fmt.Sprintf("string %d", rand.Int())
			log.Printf("generated string: %s", s)
			wg.Add(1)
			ch <- s
		}
		close(ch)
	}(wg)

	wg.Wait()
}
