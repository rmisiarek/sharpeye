package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareStrictHostFilterFunc(t *testing.T) {
	strictHostFilter, err := PrepareStrictHostFilterFunc("https://www.example.com")
	assert.Nil(t, err)

	host := "https://www.example.com"
	ok, url := strictHostFilter(&host)
	assert.Equal(t, true, ok)
	assert.Equal(t, host, url)

	host = "https://www.example.com/test"
	ok, url = strictHostFilter(&host)
	assert.Equal(t, true, ok)
	assert.Equal(t, host, url)

	host = "https://www.test.example.com/"
	ok, url = strictHostFilter(&host)
	assert.Equal(t, false, ok)
	assert.Equal(t, "", url)

	host = "bad.com"
	ok, url = strictHostFilter(&host)
	assert.Equal(t, false, ok)
	assert.Equal(t, "", url)

	strictHostFilter, err = PrepareStrictHostFilterFunc("://example.com")
	assert.NotNil(t, err)
	assert.Nil(t, strictHostFilter)
}

func TestPrepareHostFilterFunc(t *testing.T) {
	hostFilter, err := PrepareHostFilterFunc("https://www.example.com")
	assert.Nil(t, err)

	host := "https://www.example.com"
	ok, url := hostFilter(&host)
	assert.Equal(t, true, ok)
	assert.Equal(t, host, url)

	host = "https://www.example.com/test"
	ok, url = hostFilter(&host)
	assert.Equal(t, true, ok)
	assert.Equal(t, host, url)

	host = "https://www.test.example.com/"
	ok, url = hostFilter(&host)
	assert.Equal(t, true, ok)
	assert.Equal(t, host, url)

	host = "www.bad.com"
	ok, url = hostFilter(&host)
	assert.Equal(t, false, ok)
	assert.Equal(t, "", url)

	hostFilter, err = PrepareHostFilterFunc("://example.com")
	assert.NotNil(t, err)
	assert.Nil(t, hostFilter)
}
