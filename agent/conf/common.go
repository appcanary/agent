package conf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/appcanary/agent/agent/detect"
)

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

func yamlShaped(fname string) bool {
	// the only time this is not true is when it's set via command line flag
	return strings.HasSuffix(fname, ".yml")
}

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

func renameConf(path string) (err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	err = os.Rename(absPath, absPath+".deprecated")
	if err != nil {
		return
	}

	return
}

func convertOldConf() *Conf {
	env := FetchEnv()
	log := FetchLog()

	// first, reset the environment

	// TODO this is only necessary for testing, because we reach in and change
	// the value of env.ConfFile. We really should have a better way to manage
	// consts and environment fixtures. Maybe Go is just shit.
	if !strings.HasSuffix(env.ConfFile, ".conf") {
		env.ConfFile = OLD_DEFAULT_CONF_FILE
		env.VarFile = OLD_DEFAULT_VAR_FILE
	}

	// load the TOML
	c := NewTomlConfFromEnv()

	// now move the old files out of the way and dump a new YAML version
	log.Info("Old configuration detected, converting to new format")

	// For now, the /etc file can be set on the command line but the /var/db
	// file cannot. So we only have to check the ConfFile, not the VarFile
	if err := renameConf(env.ConfFile); err != nil {
		log.Warningf("Couldn't rename old agent config: %v", err)
	}

	if err := renameConf(env.VarFile); err != nil {
		log.Warningf("Couldn't rename old server config: %v", err)
	}

	// reset the new filenames
	newConfFile, err := filepath.Abs(strings.TrimSuffix(env.ConfFile, ".conf") + ".yml")
	if err != nil {
		log.Error(err)
	}

	newVarFile, err := filepath.Abs(strings.TrimSuffix(env.VarFile, ".conf") + ".yml")
	if err != nil {
		log.Error(err)
	}

	env.ConfFile = newConfFile
	env.VarFile = newVarFile

	// dump the new YAML files
	c.FullSave()

	log.Infof("New configuration file: %s", env.ConfFile)

	return c
}

func NewConfFromEnv() *Conf {
	env := FetchEnv()

	// simplest case, it's already YAML, so load and continue
	if yamlShaped(env.ConfFile) && fileExists(env.ConfFile) {
		return NewYamlConfFromEnv()
	}

	// otherwise it's a bit more complicated
	return convertOldConf()
}
