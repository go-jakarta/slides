package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctxt, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	later := time.After(3 * time.Second)
	go func() {
		defer cancel()
		for {
			select {
			case <-ch:
				log.Printf("got os signal: %d")
				return
			case <-ctxt.Done():
				log.Printf("context done: %v", ctxt.Err())
			case x := <-later:
				if !x.IsZero() {
					log.Printf("see you later")
				}
			default:
				log.Printf("nothing on any channels")
				time.Sleep(1 * time.Second)
			}
		}
	}()

	<-ch
}
