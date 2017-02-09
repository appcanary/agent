package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/appcanary/agent/agent/detect"
)

type ServerConf struct {
	UUID string `toml:"uuid" yaml:"uuid"`
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

func transformFilename(path string) string {
	log := FetchLog()

	newPath, err := filepath.Abs(strings.TrimSuffix(path, ".conf") + ".yml")
	if err != nil {
		log.Error(err)
	}

	return newPath
}

func convertOldConf() (c *Conf) {
	env := FetchEnv()
	log := FetchLog()

	// load the TOML
	if env.Prod { // we only get this far if locations are default
		c = NewTomlConfFromEnv(OLD_DEFAULT_CONF_FILE, OLD_DEFAULT_VAR_FILE)
	} else { // this should only happen in test
		log.Error("conversion of non-default config files should only happen in test")
		c = NewTomlConfFromEnv(env.ConfFile, env.VarFile)
	}

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

	var newConfFile, newVarFile string
	if env.Prod {
		newConfFile = DEFAULT_CONF_FILE
		newVarFile = DEFAULT_VAR_FILE
	} else {
		log.Error("conversion of non-default config files should only happen in test")
		newConfFile = transformFilename(env.ConfFile)
		newVarFile = transformFilename(env.VarFile)
	}

	// dump the new YAML files
	c.FullSave(newConfFile, newVarFile)

	log.Infof("New configuration file: %s", newConfFile)

	return
}

func shouldConvert(env *Env) bool {
	areTomlFiles := strings.HasSuffix(env.ConfFile, ".conf") && strings.HasSuffix(env.VarFile, ".conf")
	usingDefaults := env.ConfFile == DEFAULT_CONF_FILE && env.VarFile == DEFAULT_VAR_FILE
	notInProduction := !env.Prod
	return notInProduction || (areTomlFiles && usingDefaults)
}

func yamlFiles(env *Env) bool {
	return strings.HasSuffix(env.ConfFile, ".yml") && strings.HasSuffix(env.VarFile, ".yml")
}

func NewConfFromEnv() (c *Conf, err error) {
	env := FetchEnv()

	// simplest case, it's already YAML, so load and return
	if yamlFiles(env) {
		if fileExists(env.ConfFile) {
			c = NewYamlConfFromEnv()
		} else {
			err = errors.New(fmt.Sprintf("Couldn't find YAML conf files"))
		}
		return
	}

	// otherwise, if we're using default locations, attempt conversion
	if shouldConvert(env) {
		c = convertOldConf()
		return
	}

	err = errors.New("couldn't parse configuration file(s) - please convert to YAML")
	return
}
