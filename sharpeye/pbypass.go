package sharpeye

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fatih/color"
)

type bypassPath struct {
	statusCodeDiffer string
	pathTried        string
}

type bypassPathTarget struct {
	statusCode int // status code from base response
	target         // target
}

type pbypasser interface {
	input() chan bypassPathTarget
	run(bypassPathTarget, httper, *config, *communication)
	procesResult(result, bool)
}

type pbypass struct {
	in chan bypassPathTarget
}

func NewPathBypass() pbypass {
	return pbypass{
		in: make(chan bypassPathTarget),
	}
}

func (pb pbypass) input() chan bypassPathTarget {
	return pb.in
}

func (pb pbypass) run(t bypassPathTarget, h httper, cfg *config, msg *communication) {
	for _, payload := range cfg.Paths {
		url := fmt.Sprintf("%s://%s/%s%s",
			t.target.url.Scheme, t.target.url.Host, payload.Path, t.target.url.Path,
		)

		msg.wg.Add(1)
		go func(url, method, payload string) {
			defer msg.wg.Done()

			resp, err := h.request(url, t.method, http.Header{})
			if err != nil {
				return
			}

			b := bypassPath{pathTried: payload, statusCodeDiffer: ""}

			if resp.StatusCode != t.statusCode {
				b.statusCodeDiffer = fmt.Sprintf("%d -> %d", t.statusCode, resp.StatusCode)
			}

			msg.res <- result{
				t:          pbypassType,
				resp:       resp,
				bypassPath: b,
			}
		}(url, t.target.method, payload.Path)
	}
}

func (pb pbypass) procesResult(r result, so bool) {
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
		Info("path   | %s | %-6s | %s ( %s )", ColorStatus(r.resp.StatusCode), r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassPath.pathTried))
	} else {
		if r.bypassPath.statusCodeDiffer != "" {
			Success("path   | %s | %-6s | %s ( %s )", ColorStatus(r.resp.StatusCode), r.resp.Request.Method, r.resp.Request.URL.String(), color.GreenString(r.bypassPath.pathTried))
		} else {
			Info("path   | %s | %-6s | %s ( %s )", statusCode, r.resp.Request.Method, r.resp.Request.URL, color.BlueString(r.bypassPath.pathTried))
		}
	}
}
