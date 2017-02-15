package conf

import (
	"os"
	"testing"

	"github.com/stateio/testify/assert"
)

func TestYamlConf(t *testing.T) {
	assert := assert.New(t)

	env.ConfFile = "../test/data/test3.yml"
	env.VarFile = "../test/data/test_server3.yml"
	conf := NewYamlConfFromEnv()

	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	assert.Equal(3, len(conf.Watchers), "number of watchers")

	dpkg := conf.Watchers[0]
	assert.Equal("/var/lib/dpkg/available", dpkg.Path, "file path")

	gemfile := conf.Watchers[1]
	assert.Equal("/path/to/Gemfile.lock", gemfile.Path, "file path")

	cmd := conf.Watchers[2]
	assert.Equal("fakecmdhere", cmd.Command, "command path")

	assert.Equal("123456", conf.ServerConf.UUID)
}

func TestTomlYamlConversion(t *testing.T) {
	assert := assert.New(t)
	// need to populate DEV_CONF vars
	InitEnv("test")

	oldConfFile := OLD_DEV_CONF_FILE
	oldVarFile := OLD_DEV_VAR_FILE

	conf := NewTomlConfFromEnv(oldConfFile, oldVarFile)

	// check a few bits
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	assert.Equal("123456", conf.ServerConf.UUID)

	// now save it all as something yaml
	newConfFile := "/tmp/newagentconf.yml"
	newVarFile := "/tmp/newserverconf.yml"
	conf.FullSave(newConfFile, newVarFile)

	if _, err := os.Stat(env.ConfFile); err != nil {
		assert.Error(err)
	}

	if _, err := os.Stat(env.VarFile); err != nil {
		assert.Error(err)
	}

	// let's see what's inside
	env.ConfFile = newConfFile
	env.VarFile = newVarFile
	conf = NewYamlConfFromEnv()

	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	assert.Equal(4, len(conf.Watchers), "number of watchers")

	dpkg := conf.Watchers[0]
	assert.Equal("/var/lib/dpkg/available", dpkg.Path, "file path")

	gemfile := conf.Watchers[1]
	assert.Equal("/path/to/Gemfile.lock", gemfile.Path, "file path")

	cmd := conf.Watchers[2]
	assert.Equal("fakecmdhere", cmd.Command, "command path")

	process := conf.Watchers[3]
	assert.Equal("*", process.Process, "inspect process")

	assert.Equal("123456", conf.ServerConf.UUID)
}
