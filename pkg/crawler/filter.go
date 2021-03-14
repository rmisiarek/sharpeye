package crawler

import (
	"fmt"
	"net/url"
	"strings"
)

// PrepareStrictHostFilterFunc ...
func PrepareStrictHostFilterFunc(url string) (func(url *string) (bool, string), error) {
	hosts, err := prepareStrictHost(url)
	if err != nil {
		return nil, err
	}

	f := func(url *string) (bool, string) {
		return checkIsStrict(hosts, url)
	}

	return f, nil
}

// PrepareHostFilterFunc ...
func PrepareHostFilterFunc(url string) (func(url *string) (bool, string), error) {
	host, err := prepareHost(url)
	if err != nil {
		return nil, err
	}

	f := func(url *string) (bool, string) {
		if ok := strings.Contains(*url, host); ok {
			return true, *url
		}
		return false, ""
	}

	return f, nil
}

func checkIsStrict(hosts *[]string, url *string) (bool, string) {
	for _, p := range *hosts {
		// fmt.Printf("strings.Contains(%s, %v) -> %v \n", *url, p, strings.Contains(*url, p))
		if ok := strings.Contains(*url, p); ok {
			return true, *url
		}
	}

	return false, ""
}

func prepareStrictHost(host string) (*[]string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	host = u.Hostname()

	if strings.HasPrefix(host, "www.") {
		host = host[len("www."):]
	}

	t := []string{
		fmt.Sprintf("http://www.%s/", host),
		fmt.Sprintf("http://%s/", host),
		fmt.Sprintf("https://www.%s/", host),
		fmt.Sprintf("https://%s/", host),
	}

	return &t, nil
}

func prepareHost(host string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	host = u.Hostname()

	if strings.HasPrefix(host, "www.") {
		host = host[len("www."):]
	}

	return host, nil
}
