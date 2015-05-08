package models

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/stateio/canary-agent/agent/umwelten"
)

type Conf struct {
	ApiKey string      `toml:"api_key"`
	Files  []*FileConf `toml:"files"`
	Server *ServerConf `toml:"server"`
}

type FileConf struct {
	Path string `toml:"path"`
}

type ServerConf struct {
	UUID string `toml:"uuid"`
}

func (self *Conf) PersistServerConf(env *umwelten.Umwelten) {
	file, err := os.Create(env.VarFile)
	if err != nil {
		umwelten.Log.Fatal(err)
	}

	if err := toml.NewEncoder(file).Encode(self.Server); err != nil {
		umwelten.Log.Fatal(err)
	}

	umwelten.Log.Debug("Saved server info.")
}
