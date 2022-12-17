package main

import (
	"flag"

	"github.com/rmisiarek/sharpeye/sharpeye"
)

func main() {
	options := sharpeye.Options{}
	flag.StringVar(&options.ConfigPath, "config", "./config.yaml", "Path to YAML config file (default: ./config.yaml)")
	flag.StringVar(&options.SourcePath, "source", "./source.txt", "Path to file with URL's to test (default: ./source.txt)")
	flag.Parse()

	s, _ := sharpeye.NewSharpeye(options)
	s.Start()
}
