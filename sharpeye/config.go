package sharpeye

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var defaultConfig string = "config.yaml"

type options struct {
	source string
	config string
}

type SharpeyeConfig struct {
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

func ReadConfig() {
	dir, err := filepath.Abs(filepath.Dir("./"))
	if err != nil {
		log.Fatal(err)
	}

	yfile, err := ioutil.ReadFile(dir + "/" + defaultConfig)
	if err != nil {
		log.Fatal(err)
	}

	var data SharpeyeConfig

	err = yaml.Unmarshal(yfile, &data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("config: %v \n", data)
}
