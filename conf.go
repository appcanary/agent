package agent

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("canary-agent")

type Conf struct {
	ServerName          string    `toml:"server_name"`
	TrackSystemPackages bool      `toml:"track_system_packages"`
	LogLevel            string    `toml:"log_level"`
	Apps                []AppConf `toml:"apps"`
}

type AppConf struct {
	Name string `toml:"name"`
	Type string `toml:"type"`
	Path string `toml:"path"`
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
