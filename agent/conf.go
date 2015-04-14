package agent

import (
	"github.com/BurntSushi/toml"
	"github.com/stateio/canary-agent/agent/umwelten"
)

type Conf struct {
	ServerName          string    `toml:"server_name"`
	ApiKey              string    `toml:"api_key"`
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
		umwelten.Log.Fatal(err)
	}

	return conf
}
