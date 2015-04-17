package agent

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/stateio/canary-agent/agent/umwelten"
)

type Conf struct {
	ServerName          string      `toml:"server_name"`
	ApiKey              string      `toml:"api_key"`
	TrackSystemPackages bool        `toml:"track_system_packages"`
	LogLevel            string      `toml:"log_level"`
	Apps                []*AppConf  `toml:"apps"`
	Server              *ServerConf `toml:"server"`
}

type AppConf struct {
	Name string `toml:"name"`
	Type string `toml:"type"`
	Path string `toml:"path"`
	UUID string `toml:"uuid"`
}

type ServerConf struct {
	UUID string `toml:"uuid"`
}

func NewConf() *Conf {
	return &Conf{Server: &ServerConf{}}
}

func NewConfFromEnv() *Conf {
	conf := NewConf()

	_, err := toml.DecodeFile(env.ConfFile, &conf)
	if err != nil {
		umwelten.Log.Fatal(err)
	}

	if _, err := os.Stat(env.VarFile); err == nil {
		_, err := toml.DecodeFile(env.VarFile, &conf.Server)
		if err != nil {
			umwelten.Log.Error("%s", err)
		}
		umwelten.Log.Debug("Found, read server conf.")
	}
	return conf
}

func NewConfFromFile(path string) *Conf {
	conf := &Conf{}
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		umwelten.Log.Fatal(err)
	}

	return conf
}

func (self *Conf) PersistServerConf() {
	file, err := os.Create(env.VarFile)
	if err != nil {
		umwelten.Log.Fatal(err)
	}

	if err := toml.NewEncoder(file).Encode(self.Server); err != nil {
		umwelten.Log.Fatal(err)
	}

	umwelten.Log.Debug("Saved server info.")
}
