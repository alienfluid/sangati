package main

import (
	"encoding/json"
	"os"
)

type Test struct {
	Name     string
	Queries  []string
	Operator string
	Value    int
}

type Configuration struct {
	Host   string
	Port   int
	DbName string
	Tests  []Test
}

func parseConfigurationFile(path string, conf *Configuration) error {
	var err error

	file, err := os.Open(path)
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(conf)
	}

	return err
}
