package sharpeye

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

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
	run(bypassPathTarget, *sync.WaitGroup, httper, *config, chan<- result)
	procesResult(result)
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

func (pb pbypass) run(t bypassPathTarget, wg *sync.WaitGroup, h httper, cfg *config, r chan<- result) {
	for _, payload := range cfg.Paths {
		url := fmt.Sprintf("%s://%s%s%s",
			t.target.url.Scheme, t.target.url.Host, payload.Path, t.target.url.Path,
		)

		wg.Add(1)
		go func(url, method, payload string) {
			defer wg.Done()

			resp, err := h.request(url, t.method, http.Header{})
			if err != nil {
				return
			}

			b := bypassPath{pathTried: payload, statusCodeDiffer: ""}

			if resp.StatusCode != t.statusCode {
				b.statusCodeDiffer = fmt.Sprintf("%d -> %d", t.statusCode, resp.StatusCode)
			}

			r <- result{
				t:          pbypassType,
				resp:       resp,
				bypassPath: b,
			}
		}(url, t.target.method, payload.Path)
	}
}

func (pb pbypass) procesResult(r result) {
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
