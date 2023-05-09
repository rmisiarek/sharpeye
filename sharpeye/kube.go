package sharpeye

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fatih/color"
)

type kpresult struct {
	pathTried        string
	statusCodeDiffer string
}

type kubeProbeTarget struct {
	statusCode int // status code from base response
	target         // target
}

type kubeProber interface {
	input() chan kubeProbeTarget
	run(kubeProbeTarget, httper, *config, *communication)
	procesResult(result, bool)
}

type kubeProbe struct {
	in chan kubeProbeTarget
}

func NewKubeProbe() kubeProbe {
	return kubeProbe{
		in: make(chan kubeProbeTarget),
	}
}

func (s kubeProbe) input() chan kubeProbeTarget {
	return s.in
}

func (s kubeProbe) run(
	t kubeProbeTarget,
	h httper,
	cfg *config,
	msg *communication,
) {
	for _, path := range cfg.Kube {
		kubeUrl := fmt.Sprintf("%s://%s%s", t.target.url.Scheme, t.target.url.Host, path.Path)
		k := kpresult{pathTried: path.Path}

		msg.wg.Add(1)
		go func(url, method string) {
			defer msg.wg.Done()

			resp, err := h.request(url, method, http.Header{})
			if err != nil {
				msg.err <- err
				return
			}

			if resp.StatusCode != t.statusCode {
				k.statusCodeDiffer = fmt.Sprintf("%d -> %d", t.statusCode, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
				msg.wg.Add(1)
				go func() {
					defer msg.wg.Done()
					msg.hb <- bypassHeaderTarget{
						resp.Header, resp.StatusCode, target{url: resp.Request.URL, method: method},
					}
				}()

				if t.url.Path != "" {
					msg.wg.Add(1)
					go func() {
						defer msg.wg.Done()
						msg.pb <- bypassPathTarget{
							resp.StatusCode, target{url: resp.Request.URL, method: method},
						}
					}()
				}
			}

			msg.res <- result{t: kprobeType, resp: resp, kubeProbe: k}
		}(kubeUrl, t.method)
	}
}

func (s kubeProbe) procesResult(r result, so bool) {
	if so && (r.resp.StatusCode < 200 || r.resp.StatusCode > 300) {
		return
	}

	statusCode := strconv.Itoa(r.resp.StatusCode)
	boringCodes := []int{
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusNotFound,
		http.StatusBadRequest,
	}

	if Contains(boringCodes, r.resp.StatusCode) {
		Info("path   | %d | %-6s | %s ( %s )", r.resp.StatusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.kubeProbe.pathTried))
	} else {
		if r.kubeProbe.statusCodeDiffer != "" {
			Success("path   | %s | %-6s | %s ( %s ) [%s]", ColorStatus(r.resp.StatusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.kubeProbe.pathTried), r.kubeProbe.statusCodeDiffer)
		} else {
			Info("path   | %s | %-6s | %s ( %s ) [%s]", statusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.kubeProbe.pathTried), r.kubeProbe.statusCodeDiffer)
		}
	}
}
