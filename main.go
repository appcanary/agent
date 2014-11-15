package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	buffer, err := ioutil.ReadFile("test_gemfile.lock")
	if err != nil {
		log.Fatal(err)
	}

	gemfile := &GemfileGrammar{Buffer: string(buffer)}
	gemfile.Init()

	if err := gemfile.Parse(); err != nil {
		log.Fatal(err)
	}

	gemfile.Execute()

	for _, gem := range gemfile.Gems {
		fmt.Println(gem.Name + " : " + gem.Version)
	}

}
