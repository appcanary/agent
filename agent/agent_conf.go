package agent

import (
	"os"

	"github.com/BurntSushi/toml"
	. "github.com/appcanary/agent/agent/models"
	"github.com/appcanary/agent/agent/umwelten"
)

func NewConf() *Conf {
	return &Conf{Server: &ServerConf{}}
}

func NewConfFromEnv() *Conf {
	conf := NewConf()

	_, err := toml.DecodeFile(env.ConfFile, &conf)
	if err != nil {
		umwelten.Log.Fatal(err)
	}

	if len(conf.Files) == 0 {
		umwelten.Log.Fatal("No files to monitor!")
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
