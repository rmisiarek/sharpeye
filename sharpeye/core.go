package sharpeye

import (
	"fmt"
	"sync"
)

type resultType int

const (
	probeType resultType = iota
	bypassType
)

type result struct {
	t    resultType
	resp response
}

type communication struct {
	feedProbeCh  chan target
	feedBypassCh chan target
	resultCh     chan result
	done         chan interface{}
	wg           sync.WaitGroup
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

	fmt.Println(config)

	return sharpeye{
		client: newHttpClient(
			config.Probe.Client.Redirect,
			config.Probe.Client.Timeout,
		),
		comm: communication{
			feedProbeCh:  make(chan target),
			feedBypassCh: make(chan target),
			resultCh:     make(chan result),
			done:         make(chan interface{}),
			wg:           sync.WaitGroup{},
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
			case target := <-s.comm.feedBypassCh:
				s.bypass(target)
			case result := <-s.comm.resultCh:
				switch result.t {
				case probeType:
					processProbeResult(result)
				case bypassType:
					processBypassResult(result)
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
