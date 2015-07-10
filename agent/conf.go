package agent

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Conf struct {
	ApiKey  string      `toml:"api_key"`
	Distro  string      `toml:"distro"`
	Release string      `toml:"release"`
	Files   []*FileConf `toml:"files"`
	Server  *ServerConf `toml:"server"`
}

type FileConf struct {
	Path string `toml:"path"`
}

type ServerConf struct {
	UUID     string `toml:"uuid"`
	Hostname string `toml:"hostname"`
	Uname    string `toml:"uname"`
	Ip       string `toml:"ip"`
	Distro   string `toml:"distro"`
	Release  string `toml:"release"`
}

func NewConf() *Conf {
	return &Conf{Server: &ServerConf{}}
}

func NewConfFromEnv() *Conf {
	conf := NewConf()

	_, err := toml.DecodeFile(env.ConfFile, &conf)
	if err != nil {
		log.Fatal(err)
	}

	if len(conf.Files) == 0 {
		log.Fatal("No files to monitor!")
	}

	if _, err := os.Stat(env.VarFile); err == nil {
		_, err := toml.DecodeFile(env.VarFile, &conf.Server)
		if err != nil {
			log.Error("%s", err)
		}
		log.Debug("Found, read server conf.")
	}

	return conf
}

func (conf *Conf) PersistServerConf(env *Env) {
	file, err := os.Create(env.VarFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := toml.NewEncoder(file).Encode(conf.Server); err != nil {
		log.Fatal(err)
	}

	log.Debug("Saved server info.")
}
