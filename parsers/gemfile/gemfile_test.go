package gemfile

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	assert := assert.New(t)
	gemfile := &GemfileGrammar{Buffer: testGemfile}
	gemfile.Init()

	if err := gemfile.Parse(); err != nil {
		log.Fatal(err)
	}

	gemfile.Execute()

	assert.Equal(gemfile.Gems, testGems)
}

var testGemfile = `GEM
	remote: https://rubygems.org/
	specs:
		actionmailer (4.1.7)
			actionpack (= 4.1.7)
			actionview (= 4.1.7)
			mail (~> 2.5, >= 2.5.4)
		actionpack (4.1.7)
			actionview (= 4.1.7)
			activesupport (= 4.1.7)
			rack (~> 1.5.2)
			rack-test (~> 0.6.2)
		actionview (4.1.7)
			activesupport (= 4.1.7)
			builder (~> 3.1)
			erubis (~> 2.7.0)
		activemodel (4.1.7)
			activesupport (= 4.1.7)
			builder (~> 3.1)
		activerecord (4.1.7)
			activemodel (= 4.1.7)
			activesupport (= 4.1.7)
			arel (~> 5.0.0)
		activesupport (4.1.7)
			i18n (~> 0.6, >= 0.6.9)
			json (~> 1.7, >= 1.7.7)
			minitest (~> 5.1)
			thread_safe (~> 0.1)
			tzinfo (~> 1.1)
		arel (5.0.1.20140414130214)
		builder (3.2.2)
		erubis (2.7.0)
		hike (1.2.3)
		i18n (0.6.11)
		json (1.8.1)
		mail (2.6.3)
			mime-types (>= 1.16, < 3)
		mime-types (2.4.3)
		minitest (5.4.3)
		multi_json (1.10.1)
		rack (1.5.2)
		rack-test (0.6.2)
			rack (>= 1.0)
		rails (4.1.7)
			actionmailer (= 4.1.7)
			actionpack (= 4.1.7)
			actionview (= 4.1.7)
			activemodel (= 4.1.7)
			activerecord (= 4.1.7)
			activesupport (= 4.1.7)
			bundler (>= 1.3.0, < 2.0)
			railties (= 4.1.7)
			sprockets-rails (~> 2.0)
		railties (4.1.7)
			actionpack (= 4.1.7)
			activesupport (= 4.1.7)
			rake (>= 0.8.7)
			thor (>= 0.18.1, < 2.0)
		rake (10.3.2)
		sprockets (2.12.3)
			hike (~> 1.2)
			multi_json (~> 1.0)
			rack (~> 1.0)
			tilt (~> 1.1, != 1.3.0)
		sprockets-rails (2.2.0)
			actionpack (>= 3.0)
			activesupport (>= 3.0)
			sprockets (>= 2.8, < 4.0)
		thor (0.19.1)
		thread_safe (0.3.4)
		tilt (1.4.1)
		tzinfo (1.2.2)
			thread_safe (~> 0.1)

PLATFORMS
	ruby

DEPENDENCIES
	rails

`

var testGems = []Gem{Gem{Name: "actionmailer", Version: "(4.1.7)"},
	Gem{Name: "actionpack", Version: "(= 4.1.7)"},
	Gem{Name: "actionview", Version: "(= 4.1.7)"},
	Gem{Name: "mail", Version: "(~> 2.5, >= 2.5.4)"},
	Gem{Name: "actionpack", Version: "(4.1.7)"},
	Gem{Name: "actionview", Version: "(= 4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "rack", Version: "(~> 1.5.2)"},
	Gem{Name: "rack-test", Version: "(~> 0.6.2)"},
	Gem{Name: "actionview", Version: "(4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "builder", Version: "(~> 3.1)"},
	Gem{Name: "erubis", Version: "(~> 2.7.0)"},
	Gem{Name: "activemodel", Version: "(4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "builder", Version: "(~> 3.1)"},
	Gem{Name: "activerecord", Version: "(4.1.7)"},
	Gem{Name: "activemodel", Version: "(= 4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "arel", Version: "(~> 5.0.0)"},
	Gem{Name: "activesupport", Version: "(4.1.7)"},
	Gem{Name: "i18n", Version: "(~> 0.6, >= 0.6.9)"},
	Gem{Name: "json", Version: "(~> 1.7, >= 1.7.7)"},
	Gem{Name: "minitest", Version: "(~> 5.1)"},
	Gem{Name: "thread_safe", Version: "(~> 0.1)"},
	Gem{Name: "tzinfo", Version: "(~> 1.1)"},
	Gem{Name: "arel", Version: "(5.0.1.20140414130214)"},
	Gem{Name: "builder", Version: "(3.2.2)"},
	Gem{Name: "erubis", Version: "(2.7.0)"},
	Gem{Name: "hike", Version: "(1.2.3)"},
	Gem{Name: "i18n", Version: "(0.6.11)"},
	Gem{Name: "json", Version: "(1.8.1)"},
	Gem{Name: "mail", Version: "(2.6.3)"},
	Gem{Name: "mime-types", Version: "(>= 1.16, < 3)"},
	Gem{Name: "mime-types", Version: "(2.4.3)"},
	Gem{Name: "minitest", Version: "(5.4.3)"},
	Gem{Name: "multi_json", Version: "(1.10.1)"},
	Gem{Name: "rack", Version: "(1.5.2)"},
	Gem{Name: "rack-test", Version: "(0.6.2)"},
	Gem{Name: "rack", Version: "(>= 1.0)"},
	Gem{Name: "rails", Version: "(4.1.7)"},
	Gem{Name: "actionmailer", Version: "(= 4.1.7)"},
	Gem{Name: "actionpack", Version: "(= 4.1.7)"},
	Gem{Name: "actionview", Version: "(= 4.1.7)"},
	Gem{Name: "activemodel", Version: "(= 4.1.7)"},
	Gem{Name: "activerecord", Version: "(= 4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "bundler", Version: "(>= 1.3.0, < 2.0)"},
	Gem{Name: "railties", Version: "(= 4.1.7)"},
	Gem{Name: "sprockets-rails", Version: "(~> 2.0)"},
	Gem{Name: "railties", Version: "(4.1.7)"},
	Gem{Name: "actionpack", Version: "(= 4.1.7)"},
	Gem{Name: "activesupport", Version: "(= 4.1.7)"},
	Gem{Name: "rake", Version: "(>= 0.8.7)"},
	Gem{Name: "thor", Version: "(>= 0.18.1, < 2.0)"},
	Gem{Name: "rake", Version: "(10.3.2)"},
	Gem{Name: "sprockets", Version: "(2.12.3)"},
	Gem{Name: "hike", Version: "(~> 1.2)"},
	Gem{Name: "multi_json", Version: "(~> 1.0)"},
	Gem{Name: "rack", Version: "(~> 1.0)"},
	Gem{Name: "tilt", Version: "(~> 1.1, != 1.3.0)"},
	Gem{Name: "sprockets-rails", Version: "(2.2.0)"},
	Gem{Name: "actionpack", Version: "(>= 3.0)"},
	Gem{Name: "activesupport", Version: "(>= 3.0)"},
	Gem{Name: "sprockets", Version: "(>= 2.8, < 4.0)"},
	Gem{Name: "thor", Version: "(0.19.1)"},
	Gem{Name: "thread_safe", Version: "(0.3.4)"},
	Gem{Name: "tilt", Version: "(1.4.1)"},
	Gem{Name: "tzinfo", Version: "(1.2.2)"},
	Gem{Name: "thread_safe", Version: "(~> 0.1)"}}
