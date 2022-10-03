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
	Headers []struct {
		Header string `yaml:"header"`
	} `yaml:"headers"`
	Paths []struct {
		Path string `yaml:"path"`
	} `yaml:"paths"`
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
