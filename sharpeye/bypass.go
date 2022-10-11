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

type bypassHeader struct {
	headerTried       string
	headersReflection []string
	valuesReflection  []string
	bodyReflection    []string
	statusCodeDiffer  string
}

type bypassPath struct {
	statusCodeDiffer string
	pathTried        string
}

type bypassHeaderTarget struct {
	headers    map[string][]string // headers from base response
	statusCode int                 // status code from base response
	target                         // target
}

type bypassPathTarget struct {
	statusCode int // status code from base response
	target         // target
}

func (s *sharpeye) bypassPath(t bypassPathTarget) {
	for _, payload := range s.config.Paths {
		url := fmt.Sprintf("%s://%s%s%s",
			t.target.url.Scheme, t.target.url.Host, payload.Path, t.target.url.Path,
		)
		s.comm.wg.Add(1)
		go func(url, method, payload string) {
			defer s.comm.wg.Done()

			resp, err := s.client.request(url, t.method, http.Header{})
			if err != nil {
				return
			}

			b := bypassPath{pathTried: payload, statusCodeDiffer: ""}

			if resp.StatusCode != t.statusCode {
				b.statusCodeDiffer = fmt.Sprintf("%d -> %d", t.statusCode, resp.StatusCode)
			}

			s.comm.resultCh <- result{
				t:          bypassPathType,
				resp:       resp,
				bypassPath: b,
			}
		}(url, t.target.method, payload.Path)
	}
}

func processBypassPathResult(r result) {
	statusCode := strconv.Itoa(r.resp.StatusCode)

	switch s := statusCodeGroup(r.resp.StatusCode); s {
	case "information":
		Info("path   | %s | %-6v | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassPath.pathTried))
		Info("path   | %s | %-6v | \t>> may be interesting: %s", color.YellowString(statusCode), r.resp.Request.Method, color.YellowString(r.bypassPath.statusCodeDiffer))
	case "success":
		Success("path   | %s | %-6v | %s ( %s )", color.GreenString(statusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.bypassPath.pathTried))
		Success("path   | %s | %-6v | \t>> got it: %s", color.GreenString(statusCode), r.resp.Request.Method, color.GreenString(r.bypassPath.statusCodeDiffer))
	case "redirection":
		if r.bypassPath.statusCodeDiffer != "" && (r.resp.StatusCode != http.StatusMovedPermanently && r.resp.StatusCode != http.StatusFound) {
			Info("path   | %s | %-6v | %s ( %s )", color.YellowString(statusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.BlueString(r.bypassPath.pathTried))
			Info("path   | %s | %-6v | \t>> may be interesting: %s", color.YellowString(statusCode), r.resp.Request.Method, color.YellowString(r.bypassPath.statusCodeDiffer))
		} else {
			Info("path   | %s | %-6v | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassPath.pathTried))
		}
	case "client_error":
		if r.bypassPath.statusCodeDiffer != "" && (r.resp.StatusCode != http.StatusNotFound && r.resp.StatusCode != http.StatusBadRequest) {
			Info("path   | %s | %-6v | %s ( %s )", color.YellowString(statusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.BlueString(r.bypassPath.pathTried))
			Info("path   | %s | %-6v | \t>> may be interesting: %s", color.YellowString(statusCode), r.resp.Request.Method, color.YellowString(r.bypassPath.statusCodeDiffer))
		} else {
			Info("path   | %s | %-6v | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassPath.pathTried))
		}
	case "server_error":
		Info("path   | %s | %-6v | %s ( %s )", color.RedString(statusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.BlueString(r.bypassPath.pathTried))
		Info("path   | %s | %-6v | \t>> may be interesting: %s", color.RedString(statusCode), r.resp.Request.Method, color.RedString(r.bypassPath.statusCodeDiffer))
	}
}

func statusCodeGroup(code int) string {
	var group string
	if code >= 100 && code <= 199 {
		group = "information"
	}
	if code >= 200 && code <= 299 {
		group = "success"
	}
	if code >= 300 && code <= 399 {
		group = "redirection"
	}
	if code >= 400 && code <= 499 {
		group = "client_error"
	}
	if code >= 500 && code <= 599 {
		group = "server_error"
	}

	return group
}

func (s *sharpeye) bypassHeader(t bypassHeaderTarget) {
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

			resp, err := s.client.request(t.url.String(), t.method, http.Header{header: []string{value}})
			if err != nil {
				return
			}

			b := bypassHeader{
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
				t:            bypassHeaderType,
				resp:         resp,
				bypassHeader: b,
			}
		}(t.url.String(), t.method, header, value)
	}
}

func processBypassHeaderResult(r result) {
	delete(r.resp.Request.Header, "Connection")

	statusCode := strconv.Itoa(r.resp.StatusCode)
	if r.bypassHeader.statusCodeDiffer != "" {
		statusCode = color.GreenString(r.bypassHeader.statusCodeDiffer)
	}

	var reflectedHeaders, reflectedValues, reflectedBodyValues string

	if len(r.bypassHeader.headersReflection) > 0 {
		reflectedHeaders = strings.Join(r.bypassHeader.headersReflection, ",")
	}
	if len(r.bypassHeader.valuesReflection) > 0 {
		reflectedValues = strings.Join(r.bypassHeader.valuesReflection, ",")
	}
	if len(r.bypassHeader.bodyReflection) > 0 {
		reflectedBodyValues = strings.Join(r.bypassHeader.bodyReflection, ",")
	}

	if r.bypassHeader.statusCodeDiffer != "" || len(r.bypassHeader.headersReflection) > 0 || len(r.bypassHeader.valuesReflection) > 0 {
		Success("header | %s | %-6v | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.bypassHeader.headerTried))
		if reflectedHeaders != "" {
			Success("header | %s | %-6v | \t>> found reflected header key in headers: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedHeaders))
		}
		if reflectedValues != "" {
			Success("header | %s | %-6v | \t>> found reflected header value in headers: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedValues))
		}
		if reflectedBodyValues != "" {
			Success("header | %s | %-6v | \t>> found reflected header value in body: %s", statusCode, r.resp.Request.Method, color.GreenString(reflectedBodyValues))
		}
	} else {
		Info("header | %d | %-6v | %s ( %s )", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassHeader.headerTried))
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
