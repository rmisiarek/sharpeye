package units

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	assert.Equal(t, time.Second*30, client.Timeout)
}

func TestProbeHostWithTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TestProbeHostWithTLS")
	}))
	defer ts.Close()

	client := ts.Client()

	// Success scenario
	resp, tls, err := ProbeHost(client, ts.URL)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, tls)
	assert.Equal(t, 200, resp.StatusCode)

	// Failure scenario
	resp, tls, err = ProbeHost(client, "")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, tls)
}
