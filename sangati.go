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

func compareTime(val1 time.Time, val2 time.Time, op string) bool {
	switch {
	case op == "eq":
		return val1.Equal(val2)
	case op == "lt":
		return val1.Before(val2)
	case op == "gt":
		return val1.After(val2)
	case op == "lte":
		return val1.Before(val2) || val1.Equal(val2)
	case op == "gte":
		return val1.After(val2) || val1.Equal(val2)
	default:
		log.Fatal("Invalid operator '", op, "' specified")
	}
	return false
}

func compareInt64(val1 int64, val2 int64, op string) bool {
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

func compareString(val1 string, val2 string, op string) bool {
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
			continue
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
			if err != nil {
				log.Printf("Test '%v' FAILED. Error querying the database: %v", test.Name, err)
				failed += 1
				continue
			}

			// Validate whether the returned data is of type specified in the test
			var failure bool = false
			for i, value := range values {
				switch test.Types[i] {
				case "int":
					expected, err := strconv.Atoi(test.Values[i])
					if err != nil {
						log.Fatal("Incorrect expected value type")
					}

					if !compareInt64(value.(int64), int64(expected), test.Operator) {
						failure = true
					}
				case "string":
					expected := test.Values[i]

					b, ok := value.([]byte)
					if !ok {
						log.Fatal("Incorrect returned value type")
					}

					if !compareString(string(b), expected, test.Operator) {
						failure = true
					}
				case "date":
					expected, err := time.Parse("2006-02-01", test.Values[i])
					if err != nil {
						log.Fatal("Incorrect expected value type")
					}

					if !compareTime(value.(time.Time), expected, test.Operator) {
						failure = true
					}
				default:
					failure = true
				}

				if failure {
					break
				}
			}

			if !failure {
				log.Printf("Test '%v' PASSED", test.Name)
				succeeded += 1
			} else {
				log.Printf("Test '%v' FAILED Returned data: %v", test.Name, values)
				failed += 1
			}

		} else if len(test.Queries) == 2 {
			// Handle case where the result must be compared to the output of another query
			query1 := test.Queries[0]
			query2 := test.Queries[1]

			rows1, err := db.Query(query1)
			if err != nil {
				log.Fatal(err)
			}
			defer rows1.Close()

			rows2, err := db.Query(query2)
			if err != nil {
				log.Fatal(err)
			}
			defer rows2.Close()

			list1 := make([][]interface{}, 0)
			for rows1.Next() {
				// Make a slice for the values
				values1 := make([]interface{}, len(test.Types))
				scanArgs1 := make([]interface{}, len(values1))
				for i := range values1 {
					scanArgs1[i] = &values1[i]
				}
				err = rows1.Scan(scanArgs1...)
				if err != nil {
					log.Fatal(err)
				}
				list1 = append(list1, values1)
			}

			if err := rows1.Err(); err != nil {
				log.Fatal(err)
			}

			list2 := make([][]interface{}, 0)
			for rows2.Next() {
				// Make a slice for the values
				values2 := make([]interface{}, len(test.Types))
				scanArgs2 := make([]interface{}, len(values2))
				for i := range values2 {
					scanArgs2[i] = &values2[i]
				}
				err = rows2.Scan(scanArgs2...)
				if err != nil {
					log.Fatal(err)
				}
				list2 = append(list2, values2)
			}

			if err := rows2.Err(); err != nil {
				log.Fatal(err)
			}

			if len(list1) != len(list2) {
				log.Fatal("The two queries did not return the same number of items")
			}

			var failure bool = false
			for index, v1 := range list1 {
				v2 := list2[index]

				for i2 := range v1 {
					if !reflect.DeepEqual(v1[i2], v2[i2]) {
						failure = true
						break
					}
				}

				if failure {
					break
				}
			}

			if failure {
				log.Printf("Test '%v' FAILED", test.Name)
				failed += 1
			} else {
				log.Printf("Test '%v' SUCCEEDED", test.Name)
				succeeded += 1
			}

		} else {
			log.Fatal("Incorrect number of queries in test %s\n", test.Name)
		}
	}

	log.Printf("Total PASSED: %v, Total FAILED: %v", succeeded, failed)

}
