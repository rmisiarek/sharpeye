package sharpeye

import (
	"fmt"
	"net/http"
	"sync"
)

type resultType int

const (
	probeType resultType = iota
	bypassHeaderType
	bypassPathType
)

type result struct {
	t            resultType
	resp         *http.Response
	bypassHeader bypassHeader
	bypassPath   bypassPath
}

type communication struct {
	feedProbeCh        chan target
	feedBypassHeaderCh chan bypassHeaderTarget
	feedBypassPathCh   chan bypassPathTarget
	resultCh           chan result
	done               chan interface{}
	wg                 sync.WaitGroup
}

type sharpeye struct {
	client  *httpClient
	comm    communication
	config  config
	options Options
}

func NewSharpeye(options Options) (sharpeye, error) {
	config, err := options.loadConfig()
	if err != nil {
		panic(err)
	}

	return sharpeye{
		client: newHttpClient(
			config.Probe.Client.Redirect,
			config.Probe.Client.Timeout,
		),
		comm: communication{
			feedProbeCh:        make(chan target),
			feedBypassHeaderCh: make(chan bypassHeaderTarget),
			feedBypassPathCh:   make(chan bypassPathTarget),
			resultCh:           make(chan result),
			done:               make(chan interface{}),
			wg:                 sync.WaitGroup{},
		},
		config:  config,
		options: options,
	}, nil
}

func (s *sharpeye) startLoop() <-chan interface{} {
	ended := make(chan interface{})

	go func() {
		defer close(ended)

		for {
			select {
			case target := <-s.comm.feedProbeCh:
				s.probe(target)
			case target := <-s.comm.feedBypassHeaderCh:
				s.bypassHeader(target)
			case p := <-s.comm.feedBypassPathCh:
				s.bypassPath(p)
			case result := <-s.comm.resultCh:
				switch result.t {
				case probeType:
					processProbeResult(result)
				case bypassHeaderType:
					processBypassHeaderResult(result)
				case bypassPathType:
					processBypassPathResult(result)
				}
			case <-s.comm.done:
				fmt.Println("done!")
				return
			}
		}
	}()

	return ended
}

func (s *sharpeye) Start() {
	s.feed()
	ended := s.startLoop()

	go func() {
		s.comm.wg.Wait()
		close(s.comm.done)
	}()

	<-ended
}
