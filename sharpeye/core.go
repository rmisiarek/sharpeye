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
	client *HttpClient
	communication
	options
}

func NewSharpeye() (sharpeye, error) {
	return sharpeye{
		client: newHttpClient(false, 30),
		communication: communication{
			feedProbeCh:  make(chan target),
			feedBypassCh: make(chan target),
			resultCh:     make(chan result),
			done:         make(chan interface{}),
			wg:           sync.WaitGroup{},
		},
		options: options{
			source: "",
		},
	}, nil
}

func (s *sharpeye) startLoop() <-chan interface{} {
	ended := make(chan interface{})

	go func() {
		defer close(ended)

		for {
			select {
			case target := <-s.feedProbeCh:
				s.probe(target)
			case target := <-s.feedBypassCh:
				s.bypass(target)
			case result := <-s.resultCh:
				switch result.t {
				case probeType:
					processProbeResult(result)
				case bypassType:
					processBypassResult(result)
				}
			case <-s.done:
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
		s.wg.Wait()
		close(s.done)
	}()

	<-ended
}