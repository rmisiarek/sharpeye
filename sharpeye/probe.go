package sharpeye

import (
	"net/http"
	"sync"
)

type prober interface {
	input() chan target
	run(target, *sync.WaitGroup, httper, chan<- bypassHeaderTarget, chan<- bypassPathTarget, chan<- result)
	procesResult(result)
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
	wg *sync.WaitGroup,
	h httper,
	bh chan<- bypassHeaderTarget,
	bp chan<- bypassPathTarget,
	r chan<- result,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := h.request(t.url.String(), t.method, http.Header{})
		if err != nil {
			Error(err.Error())
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			// if resp.StatusCode == http.StatusOK {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bh <- bypassHeaderTarget{
					resp.Header, resp.StatusCode, target{url: t.url, method: t.method},
				}
			}()

			if t.url.Path != "" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					bp <- bypassPathTarget{
						resp.StatusCode, target{url: t.url, method: t.method},
					}
				}()
			}
		}

		r <- result{t: probeType, resp: resp}
	}()
}

func (s probe) procesResult(r result) {
	Info("probe | %d | %s | %v", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL)
}
