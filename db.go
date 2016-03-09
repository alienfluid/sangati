package main

import (
    "os"
    "strconv"
)

func buildConnectionString(database *Database) string {
	var connString string

	// The username and password for the db must be read from the environment
	// variables
	dbuser := os.Getenv("DBUSER" + strconv.Itoa(database.Index))
	dbpass := os.Getenv("DBPASS" + strconv.Itoa(database.Index))

	connString = "host=" + database.Host
	connString += " port=" + strconv.Itoa(database.Port)
	connString += " dbname=" + database.DbName
	connString += " user=" + dbuser
	connString += " password=" + dbpass

	return connString
}