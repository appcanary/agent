package conf

import "github.com/appcanary/agent/agent/detect"

type ServerConf struct {
	UUID string `toml:"uuid"`
}

type Conf struct {
	detect.LinuxOSInfo `yaml:",inline"`
	ApiKey             string        `yaml:"api_key,omitempty" toml:"api_key"`
	LogPath            string        `yaml:"log_path,omitempty" toml:"log_path"`
	ServerName         string        `yaml:"server_name,omitempty" toml:"server_name"`
	Watchers           []WatcherConf `yaml:"watchers" toml:"files"`
	StartupDelay       int           `yaml:"startup_delay,omitempty" toml:"startup_delay"`
	ServerConf         *ServerConf   `yaml:"-" toml:"-"`
}

type WatcherConf struct {
	Path    string `yaml:"path,omitempty" toml:"path"`
	Process string `yaml:"process,omitempty" toml:"inspect_process"`
	Command string `yaml:"command,omitempty" toml:"process"`
}

func NewConf() *Conf {
	return &Conf{ServerConf: &ServerConf{}}
}

func (c *Conf) OSInfo() *detect.LinuxOSInfo {
	if c.Distro != "" && c.Release != "" {
		return &c.LinuxOSInfo
	} else {
		return nil
	}
}
