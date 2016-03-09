package main

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"strconv"
	"time"
)

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

	dbConns := make(map[int]*sql.DB)

	for _, database := range configuration.Databases {
		connString := buildConnectionString(&database)
		db, err := sql.Open("postgres", connString)
		if err != nil {
			log.Fatal(err)
		}
		dbConns[database.Index] = db
	}

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
			dbindex := test.Queries[0].DbIndex
			query := test.Queries[0].Query

			// Make a slice for the values
			values := make([]interface{}, len(test.Values))
			scanArgs := make([]interface{}, len(values))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			err = dbConns[dbindex].QueryRow(query).Scan(scanArgs...)
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
			dbindex1 := test.Queries[0].DbIndex
			query1 := test.Queries[0].Query

			dbindex2 := test.Queries[1].DbIndex
			query2 := test.Queries[1].Query

			rows1, err := dbConns[dbindex1].Query(query1)
			if err != nil {
				log.Fatal(err)
			}
			defer rows1.Close()

			rows2, err := dbConns[dbindex2].Query(query2)
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
