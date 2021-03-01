package crawler

// TODO: write comments to exported func

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/html"
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

func (r *results) Read() {
	for {
		select {
		case data, open := <-r.data:
			if !open {
				return
			}
			fmt.Println("[data]\t", data)
		case err := <-r.err:
			fmt.Println("[error]\t", err)
		}
	}
}

func (r *results) WriteToSlice(s *[]string) {
	for {
		select {
		case data, open := <-r.data:
			if !open {
				return
			}
			*s = append(*s, data)
		case err := <-r.err:
			fmt.Println("[error]\t", err)
		}
	}
}

type crawlMode int

const (
	// ReadResults ...
	ReadResults crawlMode = iota
	// WriteResultsToSlice ...
	WriteResultsToSlice
)

// Crawl ...
func Crawl(url string, depth int, mode crawlMode) []string {
	wg := &sync.WaitGroup{}
	output := &[]string{}
	visited := newURLCache()
	results := newResults()

	go func() {
		wg.Add(1)
		go crawl(url, depth, wg, visited, results)
		wg.Wait()

		close(results.data)
	}()

	switch mode {
	case ReadResults:
		results.Read()
	case WriteResultsToSlice:
		results.WriteToSlice(output)
	}

	close(results.err)

	return *output
}

func grablUrls(resp *http.Response, node *html.Node) []string {
	var f func(*html.Node) []string
	var results []string

	f = func(n *html.Node) []string {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}
				link, err := resp.Request.URL.Parse(a.Val)
				if err != nil {
					continue
				}
				results = append(results, link.String())
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

		return results
	}

	res := f(node)
	return res
}

func crawl(url string, depth int, wg *sync.WaitGroup, cache *urlCache, res *results) {
	defer wg.Done()

	if depth == 0 || !cache.Set(url) {
		return
	}

	response, err := http.Get(url)
	if err != nil {
		res.err <- err
		return
	}
	defer response.Body.Close()

	node, err := html.Parse(response.Body)
	if err != nil {
		res.err <- err
		return
	}

	urls := grablUrls(response, node)

	res.data <- url

	for _, url := range urls {
		wg.Add(1)
		go crawl(url, depth-1, wg, cache, res)
	}
}
