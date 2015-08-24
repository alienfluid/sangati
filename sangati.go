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
		for _, query := range test.Queries {
			var cnt int
			err = db.QueryRow(query).Scan(&cnt)
			switch {
			case err == sql.ErrNoRows:
				log.Printf("No rows returned")
			case err != nil:
				log.Fatal(err)
			default:
				fmt.Printf("Count:%d\n",cnt)
			}
		}
	}

}
