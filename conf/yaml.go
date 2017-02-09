package conf

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func save(fileName string, data []byte) {
	err := ioutil.WriteFile(fileName, data, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func saveServerConf(c *Conf, varFile string) {
	log := FetchLog()

	yml, err := yaml.Marshal(c.ServerConf)
	if err != nil {
		log.Fatal(err)
	}

	save(varFile, yml)
	log.Debug("Saved server info.")
}

func (c *Conf) Save() {
	saveServerConf(c, env.VarFile)
}

// Saves the whole structure in two files
func (c *Conf) FullSave(confFile, varFile string) {
	log := FetchLog()

	yml, err := yaml.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}

	save(confFile, yml)
	saveServerConf(c, varFile)
	log.Debug("Saved all the config files.")
}

func NewYamlConfFromEnv() (conf *Conf) {
	conf = NewConf()
	log := FetchLog()
	env := FetchEnv()

	// read file contents
	data, err := ioutil.ReadFile(env.ConfFile)
	if err != nil {
		log.Error(err)
		log.Fatalf("Can't seem to read %s. Does the file exist? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile)
	}

	// parse the YAML
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		log.Error(err)
		log.Fatalf("Can't seem to parse %s. Is this file valid YAML? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile)
	}

	// bail if there's nothing configured
	if len(conf.Watchers) == 0 {
		log.Fatal("No watchers configured! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	// load the server conf (probably) from /var/db if there is one
	tryLoadingVarFile(conf)

	return conf
}

func tryLoadingVarFile(conf *Conf) {
	env := FetchEnv()
	log := FetchLog()

	if !fileExists(env.VarFile) {
		log.Debugf("%s does not exist", env.VarFile)
		return
	}

	data, err := ioutil.ReadFile(env.VarFile)
	if err != nil {
		log.Error(err)
		return
	}

	err = yaml.Unmarshal(data, &conf.ServerConf)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug("Found and read server configuration.")
}
