package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mveytsman/canary-agent/parsers/gemfile"
)

func main() {
	buffer, err := ioutil.ReadFile("gemfile_test2.txt")
	if err != nil {
		log.Fatal(err)
	}

	gf := &gemfile.GemfileParser{Buffer: string(buffer)}
	gf.Init()

	if err := gf.Parse(); err != nil {
		log.Fatal(err)
	}

	gf.Execute()

	for _, gem := range gf.Specs {
		fmt.Println(gem.Name + " : " + gem.Version)
	}

}
