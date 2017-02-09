package conf

import (
	"os"

	"github.com/BurntSushi/toml"
)

func NewTomlConfFromEnv() *Conf {
	conf := NewConf()
	log := FetchLog()
	env := FetchEnv()

	_, err := toml.DecodeFile(env.ConfFile, &conf)
	if err != nil {
		log.Fatalf("Can't seem to read %s. Does the file exist? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile)
	}

	if len(conf.Watchers) == 0 {
		log.Fatal("No files to monitor! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	if _, err := os.Stat(env.VarFile); err == nil {
		_, err := toml.DecodeFile(env.VarFile, &conf.ServerConf)
		if err != nil {
			log.Error("%s", err)
		}
		log.Debug("Found, read server conf.")
	}

	return conf
}
