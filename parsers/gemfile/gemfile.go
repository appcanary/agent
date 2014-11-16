package gemfile

type Gemfile struct {
	Specs []Spec
}

type Gem struct {
	Name    string
	Version string
}

type Spec struct {
	Gem
	Dependencies []Gem
}

func (g *Gemfile) addSpec(name string) {
	// TODO: strip out duplicate git/svn sections (https://github.com/bundler/bundler/blob/master/lib/bundler/lockfile_parser.rb#L74)
	g.Specs = append(g.Specs, Spec{Gem: Gem{Name: name}})
}

func (g *Gemfile) addSpecDep(name string) {
	last := len(g.Specs) - 1
	g.Specs[last].Dependencies = append(g.Specs[last].Dependencies, Gem{Name: name})
}

func (g *Gemfile) addSpecVersion(version string) {
	last := len(g.Specs) - 1
	g.Specs[last].Version = version
}

func (g *Gemfile) addSpecDepVersion(version string) {
	lastSpec := len(g.Specs) - 1
	lastDep := len(g.Specs[lastSpec].Dependencies)
	g.Specs[lastSpec].Dependencies[lastDep].Version = version
}
