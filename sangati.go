package main

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
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

func validateTestStructure(test *Test) bool {
	// Verify that at least one type is specified and that all types are supported
	if len(test.Types) < 1 {
		return false
	}
	for _, typ := range test.Types {
		if typ != "string" && typ != "int" && typ != "date" {
			return false
		}
	}

	// If Values are specified, the should be of the same length as the Types
	if len(test.Values) > 0 {
		if len(test.Values) != len(test.Types) {
			return false
		}
	}

	// If Values are specified, verify that they can convert to the right types
	if len(test.Values) > 0 {
		for index, typ := range test.Types {
			switch {
				case typ == "string":
					if reflect.TypeOf(test.Values[index]) != "string" {
						return false
					}
				case typ == "int":
					_, err := strconv.Atoi(test.Values[index])
					if err != nil {
						return false
					}
				case type == "date":
					_, err := time.Parse("2006-02-01", test.Values[index])
					if err != nil {
						return false
					}
				default:
					log.Fatal("Invalid type specified (error verifying values)")
			}
		}
	}

	return true
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
		log.Fatal("Could not find configuration file, Error: ", err)
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
		if len(test.Queries) == 1 {
			query := test.Queries[0]
			var cnt int
			err = db.QueryRow(query).Scan(&cnt)
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

	log.Printf("Total PASSED: %v, Total FAILED: %v", succeeded, failed)

}
