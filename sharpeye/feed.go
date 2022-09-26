package sharpeye

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
)

var HTTPMethods = []string{
	"GET",
}

type target struct {
	url    *url.URL
	method string
}

func (s *sharpeye) feed() {
	source, err := readFileOrStdin(s.options.SourcePath)
	if err != nil {
		panic(err)
	}

	defer source.Close()

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		rawURL := scanner.Text()
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			Error(err.Error())
			continue
		}
		for _, method := range s.config.Probe.Method {
			s.comm.wg.Add(1)
			go func(url *url.URL, method string) {
				defer s.comm.wg.Done()
				s.comm.feedProbeCh <- target{url: parsedURL, method: method}
			}(parsedURL, method)
		}
	}
}

func readFileOrStdin(source string) (io.ReadCloser, error) {
	r := os.Stdin

	if source != "" {
		_, err := os.Stat(source)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("[source] no such file: %s (%v)", source, err)
		}

		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("[source] can't open file: %s (%v)", source, err)
		}

		return file, nil
	}

	return r, nil
}
