package gemfile

import (
	"testing"

	"github.com/stateio/testify/assert"
)

func TestEmpty(t *testing.T) {
	// Make sure parser correctly parses empty Gemfile.lock's
	assert := assert.New(t)
	gemfile, err := ParseGemfile("test_files/Empty.Gemfile.lock")
	assert.NoError(err)
	assert.Equal([]Spec(nil), gemfile.Specs)
}

func TestRails(t *testing.T) {
	assert := assert.New(t)
	// Generic parser test using the Gemfile.lock generated when requiring rails
	gemfile, err := ParseGemfile("test_files/Rails.Gemfile.lock")
	assert.NoError(err)
	// The 'specs:' sections are parsed correctly
	for i, spec := range gemfile.Specs {
		assert.Equal(testGemfile[i].Name, spec.Name)
		assert.Equal(testGemfile[i].Version, spec.Version)
		assert.Equal(rubyGems, *spec.Source)
		assert.Equal(testGemfile[i].Dependencies, spec.Dependencies)
	}

	// The DEPENDENCIES section is parsed correctly
	assert.Equal([]Gem{Gem{Name: "rails"}}, gemfile.Dependencies)
}

var rubyGems = Source{Type: RubyGems, Options: map[string]string{"remote": "https://rubygems.org/"}}

var testGemfile = []Spec{
	Spec{Gem: Gem{Name: "actionmailer", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "actionpack", Version: "(= 4.1.7)"},
			Gem{Name: "actionview", Version: "(= 4.1.7)"},
			Gem{Name: "mail", Version: "(~> 2.5, >= 2.5.4)"}}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "actionview", Version: "(= 4.1.7)"},
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "rack", Version: "(~> 1.5.2)"},
			Gem{Name: "rack-test", Version: "(~> 0.6.2)"}}},
	Spec{Gem: Gem{Name: "actionview", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "builder", Version: "(~> 3.1)"},
			Gem{Name: "erubis", Version: "(~> 2.7.0)"}}},
	Spec{Gem: Gem{Name: "activemodel", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "builder", Version: "(~> 3.1)"}}},
	Spec{Gem: Gem{Name: "activerecord", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "activemodel", Version: "(= 4.1.7)"},
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "arel", Version: "(~> 5.0.0)"}}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "i18n", Version: "(~> 0.6, >= 0.6.9)"},
			Gem{Name: "json", Version: "(~> 1.7, >= 1.7.7)"},
			Gem{Name: "minitest", Version: "(~> 5.1)"},
			Gem{Name: "thread_safe", Version: "(~> 0.1)"},
			Gem{Name: "tzinfo", Version: "(~> 1.1)"}}},
	Spec{Gem: Gem{Name: "arel", Version: "(5.0.1.20140414130214)"}},
	Spec{Gem: Gem{Name: "builder", Version: "(3.2.2)"}},
	Spec{Gem: Gem{Name: "erubis", Version: "(2.7.0)"}},
	Spec{Gem: Gem{Name: "hike", Version: "(1.2.3)"}},
	Spec{Gem: Gem{Name: "i18n", Version: "(0.6.11)"}},
	Spec{Gem: Gem{Name: "json", Version: "(1.8.1)"}},
	Spec{Gem: Gem{Name: "mail", Version: "(2.6.3)"},
		Dependencies: []Gem{Gem{Name: "mime-types", Version: "(>= 1.16, < 3)"}}},
	Spec{Gem: Gem{Name: "mime-types", Version: "(2.4.3)"}},
	Spec{Gem: Gem{Name: "minitest", Version: "(5.4.3)"}},
	Spec{Gem: Gem{Name: "multi_json", Version: "(1.10.1)"}},
	Spec{Gem: Gem{Name: "rack", Version: "(1.5.2)"}},
	Spec{Gem: Gem{Name: "rack-test", Version: "(0.6.2)"},
		Dependencies: []Gem{Gem{Name: "rack", Version: "(>= 1.0)"}}},
	Spec{Gem: Gem{Name: "rails", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "actionmailer", Version: "(= 4.1.7)"},
			Gem{Name: "actionpack", Version: "(= 4.1.7)"},
			Gem{Name: "actionview", Version: "(= 4.1.7)"},
			Gem{Name: "activemodel", Version: "(= 4.1.7)"},
			Gem{Name: "activerecord", Version: "(= 4.1.7)"},
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "bundler", Version: "(>= 1.3.0, < 2.0)"},
			Gem{Name: "railties", Version: "(= 4.1.7)"},
			Gem{Name: "sprockets-rails", Version: "(~> 2.0)"}}},
	Spec{Gem: Gem{Name: "railties", Version: "(4.1.7)"},
		Dependencies: []Gem{
			Gem{Name: "actionpack", Version: "(= 4.1.7)"},
			Gem{Name: "activesupport", Version: "(= 4.1.7)"},
			Gem{Name: "rake", Version: "(>= 0.8.7)"},
			Gem{Name: "thor", Version: "(>= 0.18.1, < 2.0)"}}},
	Spec{Gem: Gem{Name: "rake", Version: "(10.3.2)"}},
	Spec{Gem: Gem{Name: "sprockets", Version: "(2.12.3)"},
		Dependencies: []Gem{
			Gem{Name: "hike", Version: "(~> 1.2)"},
			Gem{Name: "multi_json", Version: "(~> 1.0)"},
			Gem{Name: "rack", Version: "(~> 1.0)"},
			Gem{Name: "tilt", Version: "(~> 1.1, != 1.3.0)"}}},
	Spec{Gem: Gem{Name: "sprockets-rails", Version: "(2.2.0)"},
		Dependencies: []Gem{
			Gem{Name: "actionpack", Version: "(>= 3.0)"},
			Gem{Name: "activesupport", Version: "(>= 3.0)"},
			Gem{Name: "sprockets", Version: "(>= 2.8, < 4.0)"}}},
	Spec{Gem: Gem{Name: "thor", Version: "(0.19.1)"}},
	Spec{Gem: Gem{Name: "thread_safe", Version: "(0.3.4)"}},
	Spec{Gem: Gem{Name: "tilt", Version: "(1.4.1)"}},
	Spec{Gem: Gem{Name: "tzinfo", Version: "(1.2.2)"},
		Dependencies: []Gem{Gem{Name: "thread_safe", Version: "(~> 0.1)"}}}}
