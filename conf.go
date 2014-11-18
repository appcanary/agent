package agent

import (
	"github.com/BurntSushi/toml"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("canary-agent")

type Conf struct {
	// [database]
	Database struct {
		Location string
	}

	// [ruby]
	Ruby struct {
		Projects [][]string
	}
}

func NewConf() *Conf {
	return &Conf{}
}

func NewConfFromFile(path string) *Conf {
	conf := &Conf{}
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
