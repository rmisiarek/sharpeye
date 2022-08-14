package sharpeye

import (
	"bytes"
	"fmt"
	"net/http"
)

type bypass struct {
	bypassed bool
}

func (s *sharpeye) bypass(t target) {
	for _, payload := range s.config.Bypass[0].Payloads {

		s.comm.wg.Add(1)
		go func(url, method, header, value string) {
			defer s.comm.wg.Done()

			resp, _, err := s.client.request(t.url, t.method, http.Header{header: []string{value}})
			if err != nil {
				return
			}

			b := bypass{}
			if resp.statusCode != http.StatusUnauthorized && resp.statusCode != http.StatusForbidden {
				b = bypass{bypassed: true}
			}
			b = bypass{bypassed: false}

			s.comm.resultCh <- result{
				t:      bypassType,
				resp:   resp,
				bypass: b,
			}
		}(t.url, t.method, payload.Header, payload.Value)
	}
}

func processBypassResult(r result) {
	delete(r.resp.headers, "Connection")

	if r.bypass.bypassed {
		Success("bypass | %d | %v | %s | %s", r.resp.statusCode, r.resp.method, stringFromHeaders(r.resp.headers), r.resp.url)
	} else {
		Info("bypass | %d | %v | %s | %s", r.resp.statusCode, r.resp.method, stringFromHeaders(r.resp.headers), r.resp.url)
	}
}

func stringFromHeaders(m map[string][]string) string {
	b := new(bytes.Buffer)
	for key, values := range m {
		for _, value := range values {
			fmt.Fprintf(b, "%s: %s", key, value)
		}
	}
	return b.String()
}
