package crawler

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlCache(t *testing.T) {
	cache := newURLCache()

	_key := "test"
	exist := cache.Set(_key)
	assert.Equal(t, true, exist)

	exist = cache.Set(_key)
	assert.Equal(t, false, exist)
}

func TestCrawl(t *testing.T) {
	s := newMockServer("/crawl", baseBody)
	defer s.Close()

	// Success scenario

	url := s.URL + "/crawl"
	got := Crawl(url, 1, WriteResultsToSlice)
	assert.Equal(t, 1, len(got))

	got = Crawl(url, 1, ReadResults)
	assert.Equal(t, []string{}, got)

	// Catch loggers output

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stdout)
	}()

	// Failure scenario - unsupported protocol scheme ""

	got = Crawl("", 1, WriteResultsToSlice)
	assert.Equal(t, 0, len(got))
	assert.Contains(t, buf.String(), "unsupported protocol scheme")
	assert.Contains(t, buf.String(), "[error]")

	got = Crawl("", 1, ReadResults)
	assert.Equal(t, []string{}, got)
	assert.Contains(t, buf.String(), "unsupported protocol scheme")
	assert.Contains(t, buf.String(), "[error]")
}

func TestMakeRequest(t *testing.T) {
	s := newMockServer("/make-request", baseBody)
	defer s.Close()

	// Success scenario

	url := s.URL + "/make-request"
	resp, err := makeRequest(url)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, err)

	// Failure scenario - 404 http error

	url = s.URL + "/not-found-404"
	resp, err = makeRequest(url)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Nil(t, err)

	// Failure scenario - unsupported protocol scheme ""

	url = "very-bad-url-to-raise-error-in-GET"
	resp, err = makeRequest(url)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func TestParseBodyFail(t *testing.T) {
	// Only failure test here just because success scenario
	// is fully tested in TestMakeRequest

	dummyError := &ErrReader{errors.New("just dummy error")}

	_, err := parseBody(dummyError)
	assert.NotNil(t, err)
}

type ErrReader struct{ Error error }

func (e *ErrReader) Read([]byte) (int, error) {
	return 0, e.Error
}

func newMockServer(p string, h func(http.ResponseWriter, *http.Request)) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(p, h)

	s := httptest.NewServer(handler)

	return s
}

func baseBody(w http.ResponseWriter, r *http.Request) {
	b := []byte(`
		<!doctype html>
		<html>
		<head>
			<title>Example Domain</title>   
		</head>
		<body>
		<div>
			<h1>Example Domain</h1>
			<p><a href="https://www.example.com">Example 1</a></p>
			<p><a not-href="https://www.example.com">Example 1</a></p>
		</div>
		</body>
		</html>
	`)

	_, _ = w.Write(b)
}
