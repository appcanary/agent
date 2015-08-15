package agent

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"strings"
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
	UUID     string `toml:"uuid"`
	Hostname string `toml:"hostname"`
	Uname    string `toml:"uname"`
	Ip       string `toml:"ip"`
	Distro   string `toml:"distro"`
	Release  string `toml:"release"`
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
	if conf.Distro == "" || conf.Release == "" || conf.Distro == "unknown" || conf.Release == "unknown" {
		// We can find out distro and release on debian systems
		etcIssue, err := ioutil.ReadFile(env.DistributionFile)
		// if we fail reading, distro/os is unknown
		if err != nil {
			conf.Distro = "unknown"
			conf.Release = "unknown"
			log.Error(err.Error())
		} else {
			s := strings.Split(string(etcIssue), " ")
			conf.Distro = strings.ToLower(s[0])

			switch conf.Distro {
			case "debian":
				// /etc/issue looks like Debian GNU/Linux 8 \n \l
				conf.Release = s[2]
			case "ubuntu":
				// /etc/issue looks like Ubuntu 14.04.2 LTS \n \l
				conf.Release = s[1]
			}
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
