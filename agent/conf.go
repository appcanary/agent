package agent

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/appcanary/agent/agent/detect"
)

type Conf struct {
	ApiKey     string `toml:"api_key"`
	LogPath    string `toml:"log_path"`
	ServerName string `toml:"server_name"`
	detect.LinuxOSInfo
	Files      []*FileConf `toml:"files"`
	ServerConf *ServerConf `toml:"-"`
}

type FileConf struct {
	Path    string `toml:"path"`
	Process string `toml:"process"`
}

type ServerConf struct {
	UUID string `toml:"uuid"`
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

func (c *Conf) OSInfo() *detect.LinuxOSInfo {
	if c.Distro != "" && c.Release != "" {
		return &c.LinuxOSInfo
	} else {
		return nil
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
