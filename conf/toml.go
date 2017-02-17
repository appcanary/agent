package conf

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func NewTomlConfFromEnv(confFile, varFile string) (*Conf, error) {
	conf := NewConf()
	log := FetchLog()
	env := FetchEnv()

	_, err := toml.DecodeFile(confFile, &conf)
	if err != nil {
		log.Error(err)
		return nil, errors.New(fmt.Sprintf("Can't seem to read %s. Does the file exist? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile))
	}

	if len(conf.Watchers) == 0 {
		return nil, errors.New("No files to monitor! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	if _, err := os.Stat(varFile); err == nil {
		_, err := toml.DecodeFile(varFile, &conf.ServerConf)
		if err != nil {
			return nil, err
		}

		log.Debug("Found and read TOML server configuration")
	}

	return conf, nil
}
