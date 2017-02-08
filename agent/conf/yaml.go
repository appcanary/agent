package conf

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

func save(fileName string, data []byte) {
	err := ioutil.WriteFile(fileName, data, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Conf) Save() {
	log := FetchLog()

	yml, err := yaml.Marshal(c.ServerConf)
	if err != nil {
		log.Fatal(err)
	}

	save(env.VarFile, yml)
	log.Debug("Saved server info.")
}

// Saves the whole structure in two files
func (c *Conf) FullSave() {
	log := FetchLog()

	yml, err := yaml.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}

	save(env.ConfFile, yml)
	c.Save() // save the var file
	log.Debug("Saved all the config files.")
}

func NewYamlConfFromEnv() *Conf {
	conf := NewConf()
	log := FetchLog()
	env := FetchEnv()

	// read file contents
	data, err := ioutil.ReadFile(env.ConfFile)
	if err != nil {
		log.Fatalf("Can't seem to read %s. Does the file exist? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile)
	}

	// parse the YAML
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatalf("Can't seem to parse %s (error: %v). Is this file valid YAML? Please consult https://appcanary.com/servers/new for more instructions.", env.ConfFile, err)
	}

	// bail if there's nothing configured
	if len(conf.Watchers) == 0 {
		log.Fatal("No watchers configured! Please consult https://appcanary.com/servers/new for more instructions.")
	}

	// load the server conf from /var/db if there is one
	if _, err := os.Stat(env.VarFile); err == nil {
		data, err = ioutil.ReadFile(env.VarFile)
		if err != nil {
			log.Error(err)
		}

		err = yaml.Unmarshal(data, &conf.ServerConf)
		if err != nil {
			log.Error(err)
		}

		log.Debug("Found and read server configuration.")
	} else {
		log.Debugf("Couldn't open server configuration: %v", err)
	}

	return conf
}
