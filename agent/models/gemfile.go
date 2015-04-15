package models

import (
	"github.com/stateio/canary-agent/agent/umwelten"
	"github.com/stateio/canary-agent/parsers/gemfile"
)

type Gemfile struct {
	Path string
}

func (g *Gemfile) GetPath() string {
	return g.Path
}

func (g *Gemfile) Parse() interface{} {
	gf, err := gemfile.ParseGemfile(g.Path)
	if err != nil {
		//TODO handle error more gracefully
		//If we can't parse try again in a bit
		umwelten.Log.Fatal(err)
	}
	return gf
}
