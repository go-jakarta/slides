package main

import "log"

func main() {
	ch := make(chan error)

	go func() {
		defer func() {
			s := recover()
			log.Printf("recovered: %s", s)
		}()
		defer close(ch)

		panic("forced kill")
	}()

	<-ch
}
