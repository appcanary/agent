package conf

import (
	"os"
	"path/filepath"

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

func renameDeprecatedConf(path string) (err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	err = os.Rename(absPath, absPath+".deprecated")
	if err != nil {
		return err
	}

	return
}

func convertOldConf() (c *Conf) {
	env := FetchEnv()
	log := FetchLog()
	var conf_file, var_file string

	// load the TOML
	if env.Prod { // we only get this far if locations are default
		conf_file = OLD_DEFAULT_CONF_FILE
		var_file = OLD_DEFAULT_VAR_FILE
	} else { // this should only happen in test
		conf_file = OLD_DEV_CONF_FILE
		var_file = OLD_DEV_VAR_FILE
	}

	if fileExists(conf_file) {
		log.Info("Old configuration file detected, converting to new format")
	} else {
		// we know things are set to default AND the default yml file is missing
		// AND the old file is missing... well there's nothing for us to do here
		log.Fatal("We can't find any configuration files! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	c = NewTomlConfFromEnv(conf_file, var_file)

	// now move the old files out of the way and dump a new YAML version

	if err := renameDeprecatedConf(conf_file); err != nil {
		log.Warningf("Couldn't rename old agent config: %v", err)
	} else {
		log.Infof("Renamed %s to %s.deprecated", conf_file, conf_file)
	}

	if err := renameDeprecatedConf(var_file); err != nil {
		log.Warningf("Couldn't rename old server config: %v", err)
	} else {
		log.Infof("Renamed %s to %s.deprecated", var_file, var_file)
	}

	var newConfFile, newVarFile string
	if env.Prod {
		newConfFile = DEFAULT_CONF_FILE
		newVarFile = DEFAULT_VAR_FILE
	} else {
		newConfFile = DEV_CONF_FILE
		newVarFile = DEV_VAR_FILE
	}

	// dump the new YAML files
	c.FullSave(newConfFile, newVarFile)

	log.Infof("New configuration file saved to: %s", newConfFile)

	return c
}

func confFilesSetToDefault(env *Env) bool {
	if env.Prod {
		return env.ConfFile == DEFAULT_CONF_FILE && env.VarFile == DEFAULT_VAR_FILE
	} else {
		return env.ConfFile == DEV_CONF_FILE && env.VarFile == DEV_VAR_FILE
	}
}

// we can't function without configuration
// so at some point some substack callee of this method
// will Fatal() if it can't find what it needs
func NewConfFromEnv() (c *Conf) {
	env := FetchEnv()

	// if conf files were supplied via cli flags,
	// i.e. not the default setting,
	// they should be in yaml

	// therefore,
	// if we have a default file location
	// but the file does not exist,
	// try looking for the old files and convert them

	if confFilesSetToDefault(env) && !fileExists(env.ConfFile) {
		return convertOldConf()
	} else {
		return NewYamlConfFromEnv()
	}
}
