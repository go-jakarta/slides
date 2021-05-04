package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	urlstr := flag.String("url", "http://localhost:3000/?remote=%s", "urlstr")
	workers := flag.Int("workers", 16, "workers")
	count := flag.Int("count", 1_000_000, "count")
	delay := flag.Duration("delay", 3*time.Millisecond, "delay")
	flag.Parse()
	if err := run(context.Background(), *urlstr, *workers, *count, *delay); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, urlstr string, workers, count int, delay time.Duration) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	wg := new(sync.WaitGroup)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go do(wg, ctx, i, urlstr, r, count/workers, delay)
	}
	wg.Wait()
	return nil
}

func do(wg *sync.WaitGroup, ctx context.Context, id int, urlstr string, r *rand.Rand, total int, delay time.Duration) error {
	defer wg.Done()
	cl := &http.Client{}
	for i := 0; i < total; i++ {
		ip := fmt.Sprintf("%d.%d.%d.%d", r.Intn(255), r.Intn(255), r.Intn(255), r.Intn(255))
		req, err := http.NewRequest("GET", fmt.Sprintf(urlstr, ip), nil)
		if err != nil {
			return errf("worker %d: %v", id, err)
		}
		res, err := cl.Do(req.WithContext(ctx))
		if err != nil {
			return errf("worker %d: %v", id, err)
		}
		res.Body.Close()
		if i%100 == 0 {
			log.Printf("worker %d count: %d", id, i)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(delay):
		}
	}
	return nil
}

func errf(s string, v ...interface{}) error {
	err := fmt.Errorf(s, v...)
	log.Printf("error: %v", err)
	return err
}
