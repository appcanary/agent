package gemfile

import "strings"

type ParserState int

const (
	ParsingSpec ParserState = iota
	ParsingSpecDep
	ParsingDependency
)

type SourceType int

const (
	RubyGems SourceType = iota
	Git
	SVN
	Path
)

type Gemfile struct {
	Specs        []Spec
	Dependencies []Gem
	Sources      []Source
}

type Source struct {
	Type    SourceType
	Options map[string]string
}

type Gem struct {
	Name    string
	Version string
	Source  *Source
}

type Spec struct {
	Gem
	Dependencies []Gem
}

/*
// Defined in gemfile.peg.go
type GemfileParser struct {
	Gemfile
	ParserState

	Buffer string
	buffer []rune
	rules  [42]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}
*/

func (p *ParserState) setState(newState ParserState) {
	*p = newState
}

func (gp *GemfileParser) addGem(name string) {
	switch gp.ParserState {
	case ParsingSpec:
		gp.Gemfile.Specs = append(gp.Gemfile.Specs, Spec{Gem: Gem{Name: name}})
	case ParsingSpecDep:
		last := len(gp.Gemfile.Specs) - 1
		gp.Gemfile.Specs[last].Dependencies = append(gp.Gemfile.Specs[last].Dependencies, Gem{Name: name})
	case ParsingDependency:
		gp.Gemfile.Dependencies = append(gp.Gemfile.Dependencies, Gem{Name: name})
	}
}

func (gp *GemfileParser) addVersion(version string) {
	switch gp.ParserState {
	case ParsingSpec:
		last := len(gp.Gemfile.Specs) - 1
		gp.Gemfile.Specs[last].Version = version
	case ParsingSpecDep:
		lastSpec := len(gp.Gemfile.Specs) - 1
		lastDep := len(gp.Gemfile.Specs[lastSpec].Dependencies) - 1
		gp.Gemfile.Specs[lastSpec].Dependencies[lastDep].Version = version
	case ParsingDependency:
		last := len(gp.Gemfile.Dependencies) - 1
		gp.Gemfile.Dependencies[last].Version = version
	}
}

func (g *Gemfile) addSource(st SourceType) {
	g.Sources = append(g.Sources, Source{Type: st, Options: map[string]string{}})
}

func (g *Gemfile) addOption(option string) {
	l := strings.Split(option, ": ")
	last := len(g.Sources) - 1
	g.Sources[last].Options[l[0]] = l[1]
}
