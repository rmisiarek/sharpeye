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

type hbresult struct {
	headerTried       string
	headersReflection []string
	valuesReflection  []string
	bodyReflection    []string
	statusCodeDiffer  string
}

type bypassHeaderTarget struct {
	headers    map[string][]string // headers from base response
	statusCode int                 // status code from base response
	target                         // target
}

type hbypasser interface {
	input() chan bypassHeaderTarget
	run(bypassHeaderTarget, httper, *config, *communication)
	procesResult(result, bool)
}

type hbypass struct {
	in chan bypassHeaderTarget
}

func NewHeaderBypass() hbypass {
	return hbypass{
		in: make(chan bypassHeaderTarget),
	}
}

func (hb hbypass) input() chan bypassHeaderTarget {
	return hb.in
}

func (hb hbypass) run(t bypassHeaderTarget, h httper, cfg *config, msg *communication) {
	for _, payload := range cfg.Headers {
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

		msg.wg.Add(1)
		go func(url, method, header, value string) {
			defer msg.wg.Done()

			resp, err := h.request(t.url.String(), t.method, http.Header{header: []string{value}})
			if err != nil {
				return
			}

			b := hbresult{
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

			msg.res <- result{
				t:            hbypassType,
				resp:         resp,
				bypassHeader: b,
			}
		}(t.url.String(), t.method, header, value)
	}
}

func (hb hbypass) procesResult(r result, so bool) {
	if so && (r.resp.StatusCode < 200 || r.resp.StatusCode > 300) {
		return
	}

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
		Success("header | %s | %-6s | %s ( %s )", color.GreenString(statusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.bypassHeader.headerTried))
		if reflectedHeaders != "" {
			Success("header | %s | %-6s | \t>> found reflected header key in headers: %s", color.GreenString(statusCode), r.resp.Request.Method, color.GreenString(reflectedHeaders))
		}
		if reflectedValues != "" {
			Success("header | %s | %-6s | \t>> found reflected header value in headers: %s", color.GreenString(statusCode), r.resp.Request.Method, color.GreenString(reflectedValues))
		}
		if reflectedBodyValues != "" {
			Success("header | %s | %-6s | \t>> found reflected header value in body: %s", color.GreenString(statusCode), r.resp.Request.Method, color.GreenString(reflectedBodyValues))
		}
	} else {
		Info("header | %d | %-6s | %s ( %s )", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassHeader.headerTried))
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
