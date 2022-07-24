package sharpeye

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Options struct {
	SourcePath string
	ConfigPath string
}

type config struct {
	Probe struct {
		Client struct {
			Redirect bool `yaml:"redirect"`
			Timeout  int  `yaml:"timeout"`
		} `yaml:"client"`
		Protocol []string `yaml:"protocol"`
		Method   []string `yaml:"method"`
	} `yaml:"probe"`
	Bypass []struct {
		Name     string `yaml:"name"`
		Payloads []struct {
			Header string `yaml:"header"`
			Value  string `yaml:"value"`
		} `yaml:"payloads"`
	} `yaml:"bypass"`
}

func (o Options) loadConfig() (config, error) {
	yfile, err := ioutil.ReadFile(o.ConfigPath)
	if err != nil {
		return config{}, err
	}

	var cfg config
	err = yaml.Unmarshal(yfile, &cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}
