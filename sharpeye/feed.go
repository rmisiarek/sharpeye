package sharpeye

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

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
		rawURL := trimScheme(scanner.Text())
		for _, protocol := range s.config.Probe.Protocol {
			parsedURL, err := prepareURL(rawURL, protocol)
			if err != nil {
				Error(err.Error())
				continue
			}
			for _, method := range s.config.Probe.Method {
				s.comm.wg.Add(1)
				go func(url *url.URL, method string) {
					defer s.comm.wg.Done()
					s.probe.input() <- target{url: parsedURL, method: method}
				}(parsedURL, method)
			}
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

// prepareURL takes URL u without scheme and returns URL with s scheme
func prepareURL(u, s string) (*url.URL, error) {
	parsedURL, err := url.Parse(s + "://" + u)
	if err != nil {
		return nil, err
	}

	return parsedURL, nil
}

// trimScheme returns URL without scheme.
func trimScheme(u string) string {
	if i := strings.Index(u, "://"); i != -1 {
		return u[i+len("://"):]
	}

	return u
}
