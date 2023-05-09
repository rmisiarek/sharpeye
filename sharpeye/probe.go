package sharpeye

import (
	"net/http"
)

type prober interface {
	input() chan target
	run(target, httper, *communication)
	procesResult(result, bool)
}

type probe struct {
	in chan target
}

func NewProbe() probe {
	return probe{
		in: make(chan target),
	}
}

func (s probe) input() chan target {
	return s.in
}

func (s probe) run(
	t target,
	h httper,
	msg *communication,
) {
	msg.wg.Add(1)
	go func() {
		defer msg.wg.Done()

		resp, err := h.request(t.url.String(), t.method, http.Header{})
		if err != nil {
			msg.err <- err
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			// if resp.StatusCode == http.StatusOK {
			msg.wg.Add(1)
			go func() {
				defer msg.wg.Done()
				msg.hb <- bypassHeaderTarget{
					resp.Header, resp.StatusCode, target{url: t.url, method: t.method},
				}
			}()

			if t.url.Path != "" {
				msg.wg.Add(1)
				go func() {
					defer msg.wg.Done()
					msg.pb <- bypassPathTarget{
						resp.StatusCode, target{url: t.url, method: t.method},
					}
				}()
			}
		}

		msg.wg.Add(1)
		go func() {
			defer msg.wg.Done()
			msg.kp <- kubeProbeTarget{
				resp.StatusCode, target{url: t.url, method: t.method},
			}
		}()

		msg.res <- result{t: probeType, resp: resp}
	}()
}

func (s probe) procesResult(r result, so bool) {
	if so && (r.resp.StatusCode < 200 || r.resp.StatusCode > 300) {
		return
	}

	Info("probe  | %d | %-6s | %v", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL)
}
