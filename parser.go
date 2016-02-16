package main

import (
	"encoding/json"
	"os"
)

type Query struct {
	DbIndex	int
	Query	string
}

type Test struct {
	Name     string
	Types    []string
	Queries  []Query
	Operator string
	Values   []string
}

type Database struct {
	Host   string
	Port   int
	DbName string
	Index  int
}

type Configuration struct {
	Databases 	[]Database
	Tests  		[]Test
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
