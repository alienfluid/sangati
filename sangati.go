package main

import (
	"encoding/json"
	"os"
	"fmt"
	"strconv"
	_ "github.com/lib/pq"
	"database/sql"
	"log"
)

type Test struct {
	Name    string
	Queries	[]string
	Count	interface{}
}

type Configuration struct {
	Host	string
	Port	int
	DbName	string
	Tests	[]Test
}

func main() {
	// Get DB username and password from environment variables
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")

	// Read the configuration file
	file, err := os.Open("test.json")
	if err != nil {
		fmt.Println("error reading file:", err)
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}

	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	conn_string := "host=" + configuration.Host
	conn_string += " port=" + strconv.Itoa(configuration.Port)
	conn_string += " dbname=" + configuration.DbName
	conn_string += " user=" + dbuser
	conn_string += " password=" + dbpass

	db, err := sql.Open("postgres", conn_string)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	for _, test := range configuration.Tests {
		if len(test.Queries) == 1 {
			query := test.Queries[0]
			var cnt int
			err = db.QueryRow(query).Scan(&cnt)
			switch {
			case err != nil:
				log.Fatal(err)
			default:
				if cnt != test.Count {
					log.Fatal("Test %s failed. Expected count %d, received count %d", test.Name, test.Count, cnt)
				}
			}

		} else if len(test.Queries) == 2 {

			var op string
			switch v := test.Count.(type) {
			case string:
				op = test.Count
			default:
				log.Fatal("Invalid operator type specified")
			}
			
			query1 := test.Queries[0]
			query2 := test.Queries[1]

			var cnt1, cnt2 int

			err = db.QueryRow(query1).Scan(&cnt1)
			switch {
			case err != nil:
				log.Fatal(err)
			default:
			}

			err = db.QueryRow(query2).Scan(&cnt2)
			switch {
			case err != nil:
				log.Fatal(err)
			default:
				if 
			}

		}

		} else {
				log.Fatal("Incorrect number of queries in test %s\n", test.Name)
		}

	}

}
