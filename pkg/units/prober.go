package units

import (
	"net/http"
)

// TODO: should consider to use https://github.com/valyala/fasthttp

// ProberResponse holds data about http response.
type ProberResponse struct {
	Headers    map[string][]string
	Proto      string
	ProtoMajor int
	ProtoMinor int
	StatusCode int
}

// ProberTLSResponse holds data about TLS part of http respose.
type ProberTLSResponse struct {
	Version    uint16
	ServerName string
}

// TakeScreenshot make a request to a host given as a parameter.
// It returns ProberResponse struct with response informations
// and ProberTLSResponse struct with base TLS informations.
func ProbeHost(client *http.Client, url string) (*ProberResponse, *ProberTLSResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Add("Connection", "close")
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var tls *ProberTLSResponse
	if resp.TLS != nil {
		tls = &ProberTLSResponse{
			Version:    resp.TLS.Version,
			ServerName: resp.TLS.ServerName,
		}
	}

	// data, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	result := &ProberResponse{
		Headers:    resp.Header,
		Proto:      resp.Proto,
		ProtoMajor: resp.ProtoMajor,
		ProtoMinor: resp.ProtoMinor,
		StatusCode: resp.StatusCode,
	}

	return result, tls, nil
}

// NewHTTPClient is a constructor for http.Client
// used in the whole application
func NewHTTPClient() *http.Client {
	// var tlsConfig = &tls.Config{InsecureSkipVerify: true}

	// var dialContext = &net.Dialer{
	// 	Timeout:   time.Second * 5,
	// 	KeepAlive: time.Second,
	// }

	// var transport = &http.Transport{
	// 	MaxIdleConns:      30,
	// 	IdleConnTimeout:   time.Second,
	// 	DisableKeepAlives: true,
	// 	TLSClientConfig:   tlsConfig,
	// 	DialContext:       dialContext.DialContext,
	// }

	// re := func(req *http.Request, via []*http.Request) error {
	// 	return http.ErrUseLastResponse
	// }

	client := &http.Client{
		// Transport: transport,
		// CheckRedirect: re,
		// Timeout:       timeout,
	}

	return client
}
