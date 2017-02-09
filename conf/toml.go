package conf

import (
	"os"

	"github.com/BurntSushi/toml"
)

func NewTomlConfFromEnv(confFile, varFile string) *Conf {
	conf := NewConf()
	log := FetchLog()
	env := FetchEnv()

	_, err := toml.DecodeFile(confFile, &conf)
	if err != nil {
		log.Fatalf("Can't seem to read %s. Does the file exist? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile)
	}

	if len(conf.Watchers) == 0 {
		log.Fatal("No files to monitor! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	if _, err := os.Stat(varFile); err == nil {
		_, err := toml.DecodeFile(varFile, &conf.ServerConf)
		if err != nil {
			log.Errorf("%s", err)
		}
		log.Debug("Found and read TOML server configuration")
	}

	return conf
}
