package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	logTrace   *log.Logger
	logInfo    *log.Logger
	logWarning *log.Logger
	logError   *log.Logger
)

func logInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	logTrace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logInfo = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logWarning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logError = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Handle tests where the output from a single query is compared to a static set
// of values defined in the configuation file
func runSingleQueryTest(test Test, dbConns map[int]*sql.DB) (bool, error) {
	dbindex := test.Queries[0].DbIndex
	query := test.Queries[0].Query

	// Make a slice for the values (fields) returned from the query. We have to go through
	// this weird interface hoolabaloo because we don't know the types of the values
	// that will be returned
	values := make([]interface{}, len(test.Values))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Run the query and retrieve the results in a generic interface array
	var err error
	err = dbConns[dbindex].QueryRow(query).Scan(scanArgs...)
	if err != nil {
		return false, fmt.Errorf("Error querying the database")
	}

	// Validate whether the returned data is of types specified in the test
	for i, value := range values {
		switch test.Types[i] {
		case "int":
			expected, err := strconv.Atoi(test.Values[i])
			if err != nil {
				return false, err
			}

			if !compareInt64(value.(int64), int64(expected), test.Operator) {
				return false, fmt.Errorf("Logical constraint failed (expected= %v, returned= %v, operator= %v)", expected, value.(int64), test.Operator)
			}
		case "string":
			expected := test.Values[i]

			b, ok := value.([]byte)
			if !ok {
				return false, err
			}

			if !compareString(string(b), expected, test.Operator) {
				return false, fmt.Errorf("Logical constraint failed (expected= %v, returned= %v, operator= %v)", expected, string(b), test.Operator)
			}
		case "date":
			expected, err := time.Parse("2006-02-01", test.Values[i])
			if err != nil {
				return false, err
			}

			if !compareTime(value.(time.Time), expected, test.Operator) {
				return false, fmt.Errorf("Logical constraint failed (expected= %v, returned= %v, operator= %v)", expected, value.(time.Time), test.Operator)
			}
		default:
			return false, fmt.Errorf("Unexpected type specified (%v)", test.Types[i])
		}

	}

	return true, nil
}

// Handle tests where the output from one query is compared to the output of another
// query defined in the configuation file
func runDualQueryTest(test Test, dbConns map[int]*sql.DB) (bool, error) {
	dbindex1 := test.Queries[0].DbIndex
	query1 := test.Queries[0].Query

	dbindex2 := test.Queries[1].DbIndex
	query2 := test.Queries[1].Query

	// Execute the first query
	rows1, err := dbConns[dbindex1].Query(query1)
	if err != nil {
		return false, fmt.Errorf("Error querying the first database")
	}
	defer rows1.Close()

	// Execute the second query
	rows2, err := dbConns[dbindex2].Query(query2)
	if err != nil {
		return false, fmt.Errorf("Error querying the second database")
	}
	defer rows2.Close()

	// This type of test supports comparing a matrix of output i.e. multiple rows
	// and multiple columns can be compared to another set of rows and columns. We
	// have to thus allocate space for each row separately in a loop.
	var list1 [][]interface{}
	for rows1.Next() {
		values1 := make([]interface{}, len(test.Types))
		scanArgs1 := make([]interface{}, len(values1))
		for i := range values1 {
			scanArgs1[i] = &values1[i]
		}
		err = rows1.Scan(scanArgs1...)
		if err != nil {
			return false, fmt.Errorf("Error reading row from first query")
		}
		list1 = append(list1, values1)
	}

	if err := rows1.Err(); err != nil {
		return false, fmt.Errorf("Error reading row from first query")
	}

	var list2 [][]interface{}
	for rows2.Next() {
		values2 := make([]interface{}, len(test.Types))
		scanArgs2 := make([]interface{}, len(values2))
		for i := range values2 {
			scanArgs2[i] = &values2[i]
		}
		err = rows2.Scan(scanArgs2...)
		if err != nil {
			return false, fmt.Errorf("Error reading row from second query")
		}
		list2 = append(list2, values2)
	}

	if err := rows2.Err(); err != nil {
		return false, fmt.Errorf("Error reading row from second query")
	}

	if len(list1) != len(list2) {
		return false, fmt.Errorf("The first and the second query returned different number of rows (%v, %v)", len(list1), len(list2))
	}

	for index, v1 := range list1 {
		v2 := list2[index]

		for i2 := range v1 {
			if !reflect.DeepEqual(v1[i2], v2[i2]) {
				return false, fmt.Errorf("Values did not match at index %v", index)
			}
		}
	}

	return true, nil
}

func main() {
	var err error

	logInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// Get the path to configuation file
	confFilePtr := flag.String("c", "", "The full path to the configuration file")
	flag.Parse()

	if *confFilePtr == "" {
		log.Fatal("Configuration file not specified.")
	}

	// Read and parse the configuration file
	var configuration Configuration
	err = parseConfigurationFile(*confFilePtr, &configuration)
	if err != nil {
		log.Fatal("Could not parse configuration file, Error: ", err)
	}

	// A map of all the database connections, by index
	dbConns := make(map[int]*sql.DB)

	// Connect to each database specified in the configuation file
	for _, database := range configuration.Databases {
		connString := buildConnectionString(&database)
		db, err := sql.Open("postgres", connString)
		if err != nil {
			log.Fatal(err)
		}
		dbConns[database.Index] = db
	}

	var failed, succeeded = 0, 0

	// Run through the tests
	for _, test := range configuration.Tests {

		// First validate whether the test structure is correct
		err = validateTestStructure(&test)
		if err != nil {
			failed++
			logInfo.Printf("'%v' failed with error '%v'", test.Name, err)
			continue
		}

		var result bool
		if len(test.Queries) == 1 {
			result, err = runSingleQueryTest(test, dbConns)
			if result {
				succeeded++
				logInfo.Printf("'%v' succeeded", test.Name)
			} else {
				failed++
				logInfo.Printf("'%v' failed with error '%v'", test.Name, err)
			}
		} else if len(test.Queries) == 2 {
			result, err = runDualQueryTest(test, dbConns)
			if result {
				succeeded++
				logInfo.Printf("'%v' succeeded", test.Name)
			} else {
				failed++
				logInfo.Printf("'%v' failed with error '%v'", test.Name, err)
			}
		} else {
			logError.Fatal("Invalid configuration file format. More than two queries detected in one test.")
		}
	}

	logInfo.Printf("Total tests executed: %v", len(configuration.Tests))
	logInfo.Printf("Total succeeded: %v", succeeded)
	logInfo.Printf("Total failed: %v", failed)

}
