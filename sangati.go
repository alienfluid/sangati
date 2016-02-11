package main

import (
	"database/sql"
	"errors"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

func buildConnectionString(conf *Configuration) string {
	var connString string

	// The username and password for the db must be read from the environment
	// variables
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")

	connString = "host=" + conf.Host
	connString += " port=" + strconv.Itoa(conf.Port)
	connString += " dbname=" + conf.DbName
	connString += " user=" + dbuser
	connString += " password=" + dbpass

	return connString
}

func compareValues(val1 int, val2 int, op string) bool {
	switch {
	case op == "eq":
		return val1 == val2
	case op == "lt":
		return val1 < val2
	case op == "gt":
		return val1 > val2
	case op == "lte":
		return val1 <= val2
	case op == "gte":
		return val1 >= val2
	default:
		log.Fatal("Invalid operator '", op, "' specified")
	}
	return false
}

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
					log.Fatal("Invalid type specified (error verifying values)")
			}
		}
	}

	return nil
}

func main() {
	var err error

	// Get the path to configuation file
	confFilePtr := flag.String("c", "", "The full path to the configuration file")
	flag.Parse()

	if *confFilePtr == "" {
		log.Fatal("Configuration file not specified.")
	}

	// Read the configuration file
	var configuration Configuration
	err = parseConfigurationFile(*confFilePtr, &configuration)
	if err != nil {
		log.Fatal("Could not parse configuration file, Error: ", err)
	}

	// Build the connection string to connect to the database and then connect
	connString := buildConnectionString(&configuration)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var failed, succeeded int

	// Run through the tests
	for _, test := range configuration.Tests {
		
		// First validate whether the test structure is correct
		err = validateTestStructure(&test)
		if err != nil {
			log.Printf("Test '%v' FAILED. Invalid test structure: %v", test.Name, err)
			failed += 1
		}
		if len(test.Queries) == 1 {
			// Handle case where the result is compared to the value specified in the test
			query := test.Queries[0]
			
			// Make a slice for the values
			values := make([]interface{}, len(test.Values))
			scanArgs := make([]interface{}, len(values))
			for i := range values {
				scanArgs[i] = &values[i]
			}
			
			err = db.QueryRow(query).Scan(scanArgs...)

			// Validate whether the returned data is of type specified in the test
			for _, value := range values {
				log.Printf("%v", reflect.TypeOf(value))
					
			}
		
			/*
			switch {
			case err != nil:
				log.Fatal("Error executing test '", test.Name, "' Error:", err)
			default:
				if compareValues(cnt, test.Value, test.Operator) {
					log.Printf("Test '%v' PASSED", test.Name)
					succeeded += 1
				} else {
					log.Printf("Test '%v' FAILED. Expected %v, Received %v, Operator '%v'", test.Name, test.Value, cnt, test.Operator)
					failed += 1
				}
			}
			*/

		} else if len(test.Queries) == 2 {
			query1 := test.Queries[0]
			query2 := test.Queries[1]
			var cnt1, cnt2 int

			err = db.QueryRow(query1).Scan(&cnt1)
			switch {
			case err != nil:
				log.Fatal("Error executing test '", test.Name, "' Error:", err)
			default:
			}

			err = db.QueryRow(query2).Scan(&cnt2)
			switch {
			case err != nil:
				log.Fatal("Error executing test '", test.Name, "' Error:", err)
			default:
				if compareValues(cnt1, cnt2, test.Operator) {
					log.Printf("Test '%v' PASSED", test.Name)
					succeeded += 1
				} else {
					log.Printf("Test '%v' FAILED. Value1: %v, Value2: %v, Operator '%v'", test.Name, cnt1, cnt2, test.Operator)
					failed += 1
				}
			}

		} else {
			log.Fatal("Incorrect number of queries in test %s\n", test.Name)
		}
	}
	/*
	log.Printf("Total PASSED: %v, Total FAILED: %v", succeeded, failed)
	*/
}
