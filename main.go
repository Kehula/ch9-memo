package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Entry struct {
	res   Result
	ready chan struct{}
}

type Memo struct {
	f     Func
	mu    sync.Mutex
	cache map[string]*Entry
}

type Func func(key string) (interface{}, error)
type Result struct {
	value interface{}
	err   error
}

func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]*Entry)}
}

func (memo *Memo) Get(key string) (interface{}, error) {
	memo.mu.Lock()
	entry := memo.cache[key]
	if entry == nil {
		entry = &Entry{ready: make(chan struct{})}
		memo.cache[key] = entry
		memo.mu.Unlock()
		entry.res.value, entry.res.err = memo.f(key)

		close(entry.ready)
	} else {
		memo.mu.Unlock()
		<-entry.ready
	}

	return entry.res.value, entry.res.err
}

func main() {
	m := New(httpGetBody)
	var n sync.WaitGroup
	for _, url := range GetIncomingURLs() {
		n.Add(1)
		go func(url string) {
			defer n.Done()
			start := time.Now()
			value, err := m.Get(url)
			if err != nil {
				log.Print(err)
			}
			fmt.Printf("%s, %s, %d bytes\n", url, time.Since(start), len(value.([]byte)))
		}(url)

	}
	n.Wait()
}

func httpGetBody(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func GetIncomingURLs() []string {
	return []string{"https://golang.org",
		"https://godoc.org",
		"https://play.golang.org",
		"http://gopl.io",
		"https://golang.org",
		"https://godoc.org",
		"https://play.golang.org",
		"http://gopl.io"}
}
