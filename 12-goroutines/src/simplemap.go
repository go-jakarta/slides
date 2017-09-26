package main

import "sync"

var strs = []string{"a", "foo", "bar", "zoo", "awesome"}

func main() {
	m := struct {
		m map[string]string
		sync.RWMutex
	}{
		m: make(map[string]string),
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < len(strs); i++ {
		go func(wg *sync.WaitGroup, i int, s string) {
			m.Lock()
			defer m.Unlock()
			m.m[s] = "yes"
		}(wg, i, strs[i])
	}
	wg.Wait()
}
