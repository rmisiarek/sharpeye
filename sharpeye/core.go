package sharpeye

import (
	"fmt"
	"net/http"
	"sync"
)

type resultType int

const (
	probeType resultType = iota
	hbypassType
	pbypassType
)

type result struct {
	t            resultType
	resp         *http.Response
	bypassHeader hbresult
	bypassPath   bypassPath
}

type communication struct {
	// feedProbeCh        chan target
	feedBypassHeaderCh chan bypassHeaderTarget
	feedBypassPathCh   chan bypassPathTarget
	resultCh           chan result
	// done               chan interface{}
	wg sync.WaitGroup
}

type sharpeye struct {
	client  httper
	probe   prober
	hbypass hbypasser
	pbypass pbypasser
	comm    communication
	config  config
	options Options
}

func NewSharpeye(options Options) (sharpeye, error) {
	config, err := options.loadConfig()
	if err != nil {
		panic(err)
	}
	p := NewProbe()
	hb := NewHeaderBypass()
	pb := NewPathBypass()
	return sharpeye{
		client: NewHttpClient(
			config.Probe.Client.Redirect,
			config.Probe.Client.Timeout,
		),
		probe:   p,
		hbypass: hb,
		pbypass: pb,
		comm: communication{
			// feedProbeCh:        make(chan target),
			feedBypassHeaderCh: make(chan bypassHeaderTarget),
			feedBypassPathCh:   make(chan bypassPathTarget),
			resultCh:           make(chan result),
			// done:               make(chan interface{}),
			wg: sync.WaitGroup{},
		},
		config:  config,
		options: options,
	}, nil
}

func (s *sharpeye) startLoop(done chan interface{}) <-chan interface{} {
	ended := make(chan interface{})

	go func() {
		defer close(ended)

		for {
			select {
			case in := <-s.probe.input():
				s.probe.run(in, &s.comm.wg, s.client, s.hbypass.input(), s.pbypass.input(), s.comm.resultCh)
			case in := <-s.hbypass.input():
				s.hbypass.run(in, &s.comm.wg, s.client, &s.config, s.comm.resultCh)
			case in := <-s.pbypass.input():
				s.pbypass.run(in, &s.comm.wg, s.client, &s.config, s.comm.resultCh)
			case result := <-s.comm.resultCh:
				switch result.t {
				case probeType:
					s.probe.procesResult(result)
				case hbypassType:
					s.hbypass.procesResult(result)
				case pbypassType:
					s.pbypass.procesResult(result)
				}
			case <-done:
				fmt.Println("done!")
				return
			}
		}
	}()

	return ended
}

func (s *sharpeye) Start() {
	s.feed()
	done := make(chan interface{})
	ended := s.startLoop(done)

	go func() {
		s.comm.wg.Wait()
		close(done)
	}()

	<-ended
}
