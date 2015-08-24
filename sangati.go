package main

import (
	"encoding/json"
	"os"
	"fmt"
)

type Test struct {
	Queries	[]string
	Count	interface{}
}

type Configuration struct {
	Tests	[]Test
}

func main() {
	file, err := os.Open("src/github.com/alienfluid/sangati/test.json")
	if err != nil {
		fmt.Println("error reading file:", err)
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}

	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(configuration.Tests)
}
