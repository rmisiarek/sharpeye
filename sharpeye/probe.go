package sharpeye

import (
	"net/http"
)

func (s *sharpeye) probe(t target) {
	s.comm.wg.Add(1)

	go func() {
		defer s.comm.wg.Done()

		resp, err := s.client.request(t.url, t.method, http.Header{})
		if err != nil {
			Error(err.Error())
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			s.comm.feedBypassCh <- bypassTarget{resp.Header, resp.StatusCode, target{url: t.url, method: t.method}}
		}

		s.comm.resultCh <- result{
			t:    probeType,
			resp: resp,
		}
	}()
}

func processProbeResult(r result) {
	Info("probe  | %d | %-5s | %v", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL)
}
