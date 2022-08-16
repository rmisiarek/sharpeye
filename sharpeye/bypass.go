package sharpeye

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type bypass struct {
	headerTried       string
	headersReflection []string
	valuesReflection  []string
	statusCodeDiffer  string
}

type bypassTarget struct {
	headers    map[string][]string // headers from base response
	statusCode int                 // status code from base response
	target
}

func (s *sharpeye) bypass(t bypassTarget) {
	for _, payload := range s.config.Bypass[0].Payloads {

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

			s.comm.resultCh <- result{
				t:      bypassType,
				resp:   resp,
				bypass: b,
			}
			// rand.Seed(time.Now().UnixNano())
			// randomSleep := rand.Intn(800-200) + 200
			// time.Sleep(time.Duration(randomSleep) * time.Millisecond)
		}(t.url, t.method, payload.Header, payload.Value)
	}
}

func processBypassResult(r result) {
	delete(r.resp.Request.Header, "Connection")

	statusCode := strconv.Itoa(r.resp.StatusCode)
	if r.bypass.statusCodeDiffer != "" {
		statusCode = color.GreenString(r.bypass.statusCodeDiffer)
	}

	var reflectedHeaders, reflectedValues string
	if len(r.bypass.headersReflection) > 0 {
		reflectedHeaders = strings.Join(r.bypass.headersReflection, ",")
	}
	if len(r.bypass.valuesReflection) > 0 {
		reflectedValues = strings.Join(r.bypass.valuesReflection, ",")
	}

	reflections := color.GreenString(fmt.Sprintf("reflected keys: %s, values: %s", reflectedHeaders, reflectedValues))

	if r.bypass.statusCodeDiffer != "" || len(r.bypass.headersReflection) > 0 || len(r.bypass.valuesReflection) > 0 {
		Success("bypass | %s | %v | %s | %s", statusCode, r.resp.Request.Method, reflections, r.resp.Request.URL)
	} else {
		Info("bypass | %d | %v | %s | %s", r.resp.StatusCode, r.resp.Request.Method, r.bypass.headerTried, r.resp.Request.URL)
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
