package main

import (
	"encoding/json"
	"os"
	"fmt"
	"strconv"
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

	fmt.Println(conn_string)
}
