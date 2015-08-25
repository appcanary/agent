package agent

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type Conf struct {
	ApiKey     string      `toml:"api_key"`
	LogPath    string      `toml:"log_path"`
	Files      []*FileConf `toml:"files"`
	ServerConf *ServerConf `toml:"-"`
}

type FileConf struct {
	Path string `toml:"path"`
}

type ServerConf struct {
	UUID         string `toml:"uuid"`
	Hostname     string `toml:"hostname"`
	Uname        string `toml:"uname"`
	Ip           string `toml:"ip"`
	DistroString string `toml:"distro_string"`
}

func NewConf() *Conf {
	return &Conf{ServerConf: &ServerConf{}}
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
		_, err := toml.DecodeFile(env.VarFile, &conf.ServerConf)
		if err != nil {
			log.Error("%s", err)
		}
		log.Debug("Found, read server conf.")
	}

	return conf
}

func (conf *ServerConf) ParseDistro() {
	if conf.DistroString == "" || conf.DistroString == "unknown" {
		// We can find out distro and release on debian systems
		etcIssue, err := ioutil.ReadFile(env.DistributionFile)
		// if we fail reading, distro/os is unknown
		if err != nil {
			conf.DistroString = "unknown"
			log.Error(err.Error())
		} else {
			conf.DistroString = string(etcIssue)
		}
	}
}

func (c *Conf) Save() {
	//We actually only save the server conf
	sc := c.ServerConf
	file, err := os.Create(env.VarFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := toml.NewEncoder(file).Encode(sc); err != nil {
		log.Fatal(err)
	}

	log.Debug("Saved server info.")
}
