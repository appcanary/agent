package main

type Gemfile struct {
	Gems []Gem
}

type Gem struct {
	Name    string
	Version string
}

// Note this won't get the last gem :(
func (g *Gemfile) addGem(name string) {
	g.Gems = append(g.Gems, Gem{Name: name})

}

func (g *Gemfile) addVersion(version string) {
	g.Gems[len(g.Gems)-1].Version = version
}
