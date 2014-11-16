package gemfile

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	// Generic parser test using the Gemfile.lock generated when requiring rails
	assert := assert.New(t)

	buffer, err := ioutil.ReadFile("test_files/Rails.Gemfile.lock")
	if err != nil {
		log.Fatal(err)
	}

	gemfile := &GemfileGrammar{Buffer: string(buffer)}
	gemfile.Init()

	if err := gemfile.Parse(); err != nil {
		log.Fatal(err)
	}

	gemfile.Execute()

	assert.Equal(gemfile.Specs, railsGems)
}

var railsGems = []Spec{Spec{Gem: Gem{Name: "actionmailer", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "actionview", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "mail", Version: "(~> 2.5, >= 2.5.4)"}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "actionview", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "rack", Version: "(~> 1.5.2)"}},
	Spec{Gem: Gem{Name: "rack-test", Version: "(~> 0.6.2)"}},
	Spec{Gem: Gem{Name: "actionview", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "builder", Version: "(~> 3.1)"}},
	Spec{Gem: Gem{Name: "erubis", Version: "(~> 2.7.0)"}},
	Spec{Gem: Gem{Name: "activemodel", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "builder", Version: "(~> 3.1)"}},
	Spec{Gem: Gem{Name: "activerecord", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "activemodel", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "arel", Version: "(~> 5.0.0)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "i18n", Version: "(~> 0.6, >= 0.6.9)"}},
	Spec{Gem: Gem{Name: "json", Version: "(~> 1.7, >= 1.7.7)"}},
	Spec{Gem: Gem{Name: "minitest", Version: "(~> 5.1)"}},
	Spec{Gem: Gem{Name: "thread_safe", Version: "(~> 0.1)"}},
	Spec{Gem: Gem{Name: "tzinfo", Version: "(~> 1.1)"}},
	Spec{Gem: Gem{Name: "arel", Version: "(5.0.1.20140414130214)"}},
	Spec{Gem: Gem{Name: "builder", Version: "(3.2.2)"}},
	Spec{Gem: Gem{Name: "erubis", Version: "(2.7.0)"}},
	Spec{Gem: Gem{Name: "hike", Version: "(1.2.3)"}},
	Spec{Gem: Gem{Name: "i18n", Version: "(0.6.11)"}},
	Spec{Gem: Gem{Name: "json", Version: "(1.8.1)"}},
	Spec{Gem: Gem{Name: "mail", Version: "(2.6.3)"}},
	Spec{Gem: Gem{Name: "mime-types", Version: "(>= 1.16, < 3)"}},
	Spec{Gem: Gem{Name: "mime-types", Version: "(2.4.3)"}},
	Spec{Gem: Gem{Name: "minitest", Version: "(5.4.3)"}},
	Spec{Gem: Gem{Name: "multi_json", Version: "(1.10.1)"}},
	Spec{Gem: Gem{Name: "rack", Version: "(1.5.2)"}},
	Spec{Gem: Gem{Name: "rack-test", Version: "(0.6.2)"}},
	Spec{Gem: Gem{Name: "rack", Version: "(>= 1.0)"}},
	Spec{Gem: Gem{Name: "rails", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "actionmailer", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "actionview", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activemodel", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activerecord", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "bundler", Version: "(>= 1.3.0, < 2.0)"}},
	Spec{Gem: Gem{Name: "railties", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "sprockets-rails", Version: "(~> 2.0)"}},
	Spec{Gem: Gem{Name: "railties", Version: "(4.1.7)"}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(= 4.1.7)"}},
	Spec{Gem: Gem{Name: "rake", Version: "(>= 0.8.7)"}},
	Spec{Gem: Gem{Name: "thor", Version: "(>= 0.18.1, < 2.0)"}},
	Spec{Gem: Gem{Name: "rake", Version: "(10.3.2)"}},
	Spec{Gem: Gem{Name: "sprockets", Version: "(2.12.3)"}},
	Spec{Gem: Gem{Name: "hike", Version: "(~> 1.2)"}},
	Spec{Gem: Gem{Name: "multi_json", Version: "(~> 1.0)"}},
	Spec{Gem: Gem{Name: "rack", Version: "(~> 1.0)"}},
	Spec{Gem: Gem{Name: "tilt", Version: "(~> 1.1, != 1.3.0)"}},
	Spec{Gem: Gem{Name: "sprockets-rails", Version: "(2.2.0)"}},
	Spec{Gem: Gem{Name: "actionpack", Version: "(>= 3.0)"}},
	Spec{Gem: Gem{Name: "activesupport", Version: "(>= 3.0)"}},
	Spec{Gem: Gem{Name: "sprockets", Version: "(>= 2.8, < 4.0)"}},
	Spec{Gem: Gem{Name: "thor", Version: "(0.19.1)"}},
	Spec{Gem: Gem{Name: "thread_safe", Version: "(0.3.4)"}},
	Spec{Gem: Gem{Name: "tilt", Version: "(1.4.1)"}},
	Spec{Gem: Gem{Name: "tzinfo", Version: "(1.2.2)"}},
	Spec{Gem: Gem{Name: "thread_safe", Version: "(~> 0.1)"}}}
