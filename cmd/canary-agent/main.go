package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mveytsman/canary-agent/parsers/gemfile"
)

func main() {
	buffer, err := ioutil.ReadFile("test_gemfile.lock")
	if err != nil {
		log.Fatal(err)
	}

	gf := &gemfile.GemfileGrammar{Buffer: string(buffer)}
	gf.Init()

	if err := gf.Parse(); err != nil {
		log.Fatal(err)
	}

	gf.Execute()

	for _, gem := range gf.Gems {
		fmt.Println(gem.Name + " : " + gem.Version)
	}

}
