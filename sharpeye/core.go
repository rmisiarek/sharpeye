package sharpeye

import (
	"fmt"
	"net/http"
	"sync"
)

type resultType int

const (
	probeType resultType = iota
	kprobeType
	hbypassType
	pbypassType
)

type result struct {
	t            resultType
	resp         *http.Response
	kubeProbe    kpresult
	bypassHeader hbresult
	bypassPath   bypassPath
}

type communication struct {
	// feedProbeCh        chan target
	kp  chan kubeProbeTarget
	hb  chan bypassHeaderTarget
	pb  chan bypassPathTarget
	res chan result
	err chan error
	// done               chan interface{}
	wg sync.WaitGroup
}

type sharpeye struct {
	client  httper
	probe   prober
	kprobe  kubeProber
	hbypass hbypasser
	pbypass pbypasser
	msg     communication
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
	kp := NewKubeProbe()
	return sharpeye{
		client: NewHttpClient(
			config.Probe.Client.Redirect,
			config.Probe.Client.Timeout,
		),
		probe:   p,
		kprobe:  kp,
		hbypass: hb,
		pbypass: pb,
		msg: communication{
			// feedProbeCh:        make(chan target),
			kp:  kp.input(),
			hb:  hb.input(),
			pb:  pb.input(),
			res: make(chan result),
			err: make(chan error),
			wg:  sync.WaitGroup{},
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
				s.probe.run(in, s.client, &s.msg)
			case in := <-s.kprobe.input():
				s.kprobe.run(in, s.client, &s.config, &s.msg)
			case in := <-s.hbypass.input():
				s.hbypass.run(in, s.client, &s.config, &s.msg)
			case in := <-s.pbypass.input():
				s.pbypass.run(in, s.client, &s.config, &s.msg)
			case in := <-s.msg.err:
				processError(in, s.config.Probe.ShowErrors)
			case result := <-s.msg.res:
				switch result.t {
				case probeType:
					s.probe.procesResult(result, s.config.Probe.SuccessOnly)
				case kprobeType:
					s.kprobe.procesResult(result, s.config.Probe.SuccessOnly)
				case hbypassType:
					s.hbypass.procesResult(result, s.config.Probe.SuccessOnly)
				case pbypassType:
					s.pbypass.procesResult(result, s.config.Probe.SuccessOnly)
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
		s.msg.wg.Wait()
		close(done)
	}()

	<-ended
}
