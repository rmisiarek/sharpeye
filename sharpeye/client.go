package sharpeye

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type httpClient struct {
	client *http.Client
}

func newHttpClient(followRedirect bool, timeout int) *httpClient {
	var tlsConfig = &tls.Config{InsecureSkipVerify: true}

	var dialContext = &net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second,
	}

	var transport = &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   tlsConfig,
		DialContext:       dialContext.DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	if followRedirect {
		client.CheckRedirect = nil
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &httpClient{client: client}
}

func (h *httpClient) request(url string, method string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) != 0 {
		req.Header = headers
	}

	req.Header.Add("Connection", "close")
	req.Close = true

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}
