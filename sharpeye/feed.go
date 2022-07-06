package sharpeye

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var HTTPMethods = []string{
	"GET",
}

type target struct {
	url    string
	method string
}

func (s *sharpeye) feed() {
	source, err := readFileOrStdin(s.options.source)
	if err != nil {
		panic(err)
	}

	defer source.Close()

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		for _, method := range HTTPMethods {
			s.wg.Add(1)
			m := method
			go func(url string) {
				defer s.wg.Done()
				s.feedProbeCh <- target{url: url, method: m}
			}(scanner.Text())
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
