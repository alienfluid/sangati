package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// Query describes the SQL query to be executed along with the index of the
// database on which it should be executed
type Query struct {
	DbIndex int
	Query   string
}

// Test describes a unit of test that needs to be executed
type Test struct {
	Name     string
	Types    []string
	Queries  []Query
	Operator string
	Values   []string
}

// Database describes the connection details and index of the database to connect to
type Database struct {
	Host   string
	Port   int
	DbName string
	Index  int
}

// Configuration describes the set of databases and tests that needs to run
type Configuration struct {
	Databases []Database
	Tests     []Test
}

// Parses the specified configuration file to make sure it adheres to the structure
// we expect
func parseConfigurationFile(path string, conf *Configuration) error {
	var err error

	file, err := os.Open(path)
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(conf)
	}

	return err
}

// Validates individual tests to make sure they have the right set of parameters specified
func validateTestStructure(test *Test) error {
	// Verify that at least one type is specified and that all types are supported
	if len(test.Types) < 1 {
		return errors.New("The Types array must contain at least one supported type")
	}
	for _, typ := range test.Types {
		if typ != "string" && typ != "int" && typ != "date" {
			return errors.New("The Types array contains an unsupported type")
		}
	}

	// If Values are specified, the should be of the same length as the Types
	if len(test.Values) > 0 {
		if len(test.Values) != len(test.Types) {
			return errors.New("The Types array and the Values array are not of the same length")
		}
	}

	// If Values are specified, verify that they can convert to the right types
	if len(test.Values) > 0 {
		for index, typ := range test.Types {
			switch {
			case typ == "string":
				if reflect.TypeOf(test.Values[index]) != reflect.TypeOf(" ") {
					return errors.New("Value of type string expected")
				}
			case typ == "int":
				_, err := strconv.Atoi(test.Values[index])
				if err != nil {
					return errors.New("Value of type int expected")
				}
			case typ == "date":
				_, err := time.Parse("2006-02-01", test.Values[index])
				if err != nil {
					return errors.New("Value of type date (YYYY-MM-DD) expected")
				}
			default:
				return fmt.Errorf("Unsupported value type specified (%v)", typ)
			}
		}
	}

	return nil
}
