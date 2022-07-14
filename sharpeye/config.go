package sharpeye

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type SharpeyeOptions struct {
	SourcePath string
	ConfigPath string
	Config     sharpeyeConfig
}

type sharpeyeConfig struct {
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
			Payload string `yaml:"payload"`
			Value   string `yaml:"value"`
		} `yaml:"payloads"`
	} `yaml:"bypass"`
}

func (o SharpeyeOptions) loadConfig() error {
	yfile, err := ioutil.ReadFile(o.ConfigPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yfile, &o.Config)
	if err != nil {
		return err
	}

	return nil
}
