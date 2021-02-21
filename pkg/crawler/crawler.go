package crawler

import (
	"sync"
)

type urlCache struct {
	urls map[string]struct{}
	sync.Mutex
}

func (v *urlCache) Set(url string) bool {
	v.Lock()
	defer v.Unlock()

	_, exist := v.urls[url]
	v.urls[url] = struct{}{}

	return !exist
}

func newURLCache() *urlCache {
	return &urlCache{
		urls: make(map[string]struct{}),
	}
}

type results struct {
	data chan string
	err  chan error
}

func newResults() *results {
	return &results{
		data: make(chan string, 1),
		err:  make(chan error, 1),
	}
}
