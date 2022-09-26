package sharpeye

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type bypass struct {
	headerTried       string
	headersReflection []string
	valuesReflection  []string
	bodyReflection    []string
	statusCodeDiffer  string
}

type bypassTarget struct {
	headers    map[string][]string // headers from base response
	statusCode int                 // status code from base response
	target
}

func (s *sharpeye) bypass(t bypassTarget) {
	for _, payload := range s.config.Headers {

		var header, value string
		p := strings.Split(strings.TrimSpace(payload.Header), ":")

		switch s := len(p); {
		case s == 2:
			header = p[0]
			value = p[1]
		case s == 1:
			header = p[0]
			value = "127.0.0.1, 0.0.0.0, localhost"
		default:
			continue
		}

		s.comm.wg.Add(1)
		go func(url, method, header, value string) {
			defer s.comm.wg.Done()

			resp, err := s.client.request(t.url, t.method, http.Header{header: []string{value}})
			if err != nil {
				return
			}

			b := bypass{
				headerTried: fmt.Sprintf("%s: %s", header, value),
			}

			if resp.StatusCode != t.statusCode {
				b.statusCodeDiffer = fmt.Sprintf("%d -> %d", t.statusCode, resp.StatusCode)
			}

			headers, values := normalizeHeaders(resp.Header)

			foundHeaders, ok := searchForReflection(header, headers)
			if ok {
				b.headersReflection = foundHeaders
			}

			foundValues, ok := searchForReflection(value, values)
			if ok {
				b.valuesReflection = foundValues
			}

			foundBodyReflection, ok := reflectionInBody(resp.Body, values)
			if ok {
				b.bodyReflection = foundBodyReflection
			}

			s.comm.resultCh <- result{
				t:      bypassType,
				resp:   resp,
				bypass: b,
			}
		}(t.url, t.method, header, value)
	}
}

func processBypassResult(r result) {
	delete(r.resp.Request.Header, "Connection")

	statusCode := strconv.Itoa(r.resp.StatusCode)
	if r.bypass.statusCodeDiffer != "" {
		statusCode = color.GreenString(r.bypass.statusCodeDiffer)
	}

	var reflectedHeaders, reflectedValues, reflectedBodyValues string

	if len(r.bypass.headersReflection) > 0 {
		reflectedHeaders = strings.Join(r.bypass.headersReflection, ",")
	}
	if len(r.bypass.valuesReflection) > 0 {
		reflectedValues = strings.Join(r.bypass.valuesReflection, ",")
	}
	if len(r.bypass.bodyReflection) > 0 {
		reflectedBodyValues = strings.Join(r.bypass.bodyReflection, ",")
	}

	if r.bypass.statusCodeDiffer != "" || len(r.bypass.headersReflection) > 0 || len(r.bypass.valuesReflection) > 0 {
		Success("bypass | %s | %-6v | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.bypass.headerTried))
		if reflectedHeaders != "" {
			Success("bypass | %s | %-6v | \t>> found reflected header key in headers: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedHeaders))
		}
		if reflectedValues != "" {
			Success("bypass | %s | %-6v | \t>> found reflected header value in headers: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedValues))
		}
		if reflectedBodyValues != "" {
			Success("bypass | %s | %-6v | \t>> found reflected header value in body: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedBodyValues))
		}
	} else {
		Info("bypass | %d | %-6v | %s ( %s )", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypass.headerTried))
	}
}

func normalizeHeaders(response http.Header) ([]string, []string) {
	headers := []string{}
	values := []string{}

	for k, vals := range response {
		headers = append(headers, strings.ToLower(k))
		for _, v := range vals {
			values = append(values, strings.ToLower(v))
		}
	}

	return headers, values
}

func searchForReflection(value string, values []string) ([]string, bool) {
	results := []string{}
	var found bool

	for _, v := range values {
		if strings.EqualFold(value, v) {
			found = true
			results = append(results, v)
		}
	}

	return results, found
}

func reflectionInBody(response io.Reader, values []string) ([]string, bool) {
	results := []string{}
	var found bool

	body, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, false
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(string(body)), strings.ToLower(value)) {
			found = true
			results = append(results, value)
		}
	}

	return results, found
}
