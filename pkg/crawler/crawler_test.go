package crawler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlCache(t *testing.T) {
	cache := newURLCache()

	// First use - there is no 'test' in cache, so true is returned
	_key := "test"
	exist := cache.Set(_key)
	assert.Equal(t, true, exist)

	// Second use - now 'test' exist in cache, so false is returned
	exist = cache.Set(_key)
	assert.Equal(t, false, exist)
}

func TestCrawl(t *testing.T) {
	s := newMockServer("/crawl", baseBody)
	defer s.Close()

	// Crawler initialization
	c := NewCrawler()

	// Simple custom function to read results
	result := []string{}
	resultDataF := func(data string) {
		result = append(result, data)
	}

	// Simple custom function to test error
	errors := []error{}
	resultErrorF := func(err error) {
		errors = append(errors, err)
	}

	// Simple custom function to test error
	filterURL := func(url *string) (bool, string) {
		if strings.Contains(*url, "example") {
			return true, *url
		}
		return false, ""
	}

	// Set custom functions
	c.SetResultData(resultDataF)
	c.SetResultError(resultErrorF)

	// Success scenario
	// result[0] is a s.URL -> because of that below I assume there will be 3 items in results
	c.Crawl(s.URL+"/crawl", 2)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, true, isInSlice(result, "https://www.example.com"))
	assert.Equal(t, true, isInSlice(result, "https://www.test.com"))

	// Set custom filter function
	// Now only one URL should be found - www.example.com
	result = []string{}
	c.SetFilterURL(filterURL)
	c.Crawl(s.URL+"/crawl", 2)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, true, isInSlice(result, "https://www.example.com"))
	assert.Equal(t, false, isInSlice(result, "https://www.test.com"))

	// Failure scenario - unsupported protocol scheme ""
	c.Crawl("", 1)
	assert.Equal(t, 1, len(errors))
	assert.NotNil(t, errors[0])

	// Catch loggers output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stdout)
	}()

	// Test default defaultResultData
	c = NewCrawler()
	c.Crawl(s.URL+"/crawl", 2)
	assert.Contains(t, buf.String(), "www.example.com")
	assert.Contains(t, buf.String(), "www.test.com")

	// Test default defaultResultError - unsupported protocol scheme ""
	c.Crawl("", 1)
	assert.Contains(t, buf.String(), "unsupported protocol scheme")
}

var simpleResponseBody = []byte(`
	<!doctype html>
	<html>
	<head>
		<title>Example Domain</title>
	</head>
	<body>
	<div>
		<h1>Example Domain</h1>
		<p><a href="https://www.example.com">This URL should be found</a></p>
		<p><a href="https://www.test.com">This should not be found if filtered is OK</a></p>
		<p><a not-href="https://www.not-valid-href.com">It's not a valid 'a href'</a></p>
	</div>
	</body>
	</html>
`)

func newMockServer(p string, h func(http.ResponseWriter, *http.Request)) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(p, h)

	s := httptest.NewServer(handler)

	return s
}

func baseBody(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(simpleResponseBody))
}

func isInSlice(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
