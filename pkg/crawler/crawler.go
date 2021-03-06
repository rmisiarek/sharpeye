package crawler

import (
	"io"
	"log"
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

type filterURL func(url *string) (bool, string)

type resultData func(data string)

type resultError func(err error)

// Crawler it's a main package struct which hold user filter functions
// and all necessary fields for recursive crawling.
// One should use NewCrawler constructor to get one.
type Crawler struct {
	URL          string
	Depth        int
	filterURL    filterURL
	resultDataF  resultData
	resultErrorF resultError
	wg           *sync.WaitGroup
	resultData   chan string
	resultErr    chan error
}

// NewCrawler it's a constructor for Crawler structure.
// It returns instance with default functions
// which can be changed with usage of Set* methods.
func NewCrawler() *Crawler {
	return &Crawler{
		filterURL:    defaultFilterURL,
		resultDataF:  defaultResultData,
		resultErrorF: defaultResultError,
	}
}

// SetFilterURL enables to set custom function for URL filtering.
func (c *Crawler) SetFilterURL(f filterURL) *Crawler {
	c.filterURL = f
	return c
}

// SetResultData enables to set custom function for managing
// data returned by main Crawler function.
func (c *Crawler) SetResultData(f resultData) *Crawler {
	c.resultDataF = f
	return c
}

// SetResultError enables to set custom function for managing
// errors returned by main Crawler function.
func (c *Crawler) SetResultError(f resultError) *Crawler {
	c.resultErrorF = f
	return c
}

// Crawl it's a main Crawler function which will start crawling process.
// Parameters:
//		url: URL for crawling
//		depth: after exceeding the action will be stopped
func (c *Crawler) Crawl(url string, depth int) []string {
	output := &[]string{}
	visited := newURLCache()

	c.resultData = make(chan string, 1)
	c.resultErr = make(chan error, 1)
	c.wg = &sync.WaitGroup{}

	go func() {
		c.wg.Add(1)
		go c.crawl(url, depth, visited)
		c.wg.Wait()

		close(c.resultData)
	}()

	c.readResult()
	close(c.resultErr)

	return *output
}

func (c *Crawler) readResult() {
	for {
		select {
		case data, open := <-c.resultData:
			if !open {
				return
			}
			c.resultDataF(data)
		case err := <-c.resultErr:
			c.resultErrorF(err)
		}
	}
}

func (c *Crawler) crawl(url string, depth int, cache *urlCache) {

	defer c.wg.Done()

	if depth == 0 || !cache.Set(url) {
		return
	}

	response, err := makeRequest(url)
	if err != nil {
		c.resultErr <- err
		return
	}
	defer response.Body.Close()

	node, err := parseBody(response.Body)
	if err != nil {
		c.resultErr <- err
		return
	}

	urls := grablUrls(response, node)

	ok, filteredURL := c.filterURL(&url)
	if ok {
		c.resultData <- filteredURL
	}

	for _, url := range urls {
		c.wg.Add(1)
		go c.crawl(url, depth-1, cache)
	}
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

func makeRequest(url string) (r *http.Response, err error) {
	// TODO: use customized http.Client

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func parseBody(r io.Reader) (*html.Node, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func defaultFilterURL(url *string) (bool, string) {
	return true, *url
}

func defaultResultData(data string) {
	log.Println(data)
}

func defaultResultError(err error) {
	log.Println(err)
}
