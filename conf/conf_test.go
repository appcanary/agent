package conf

import (
	"os"
	"strings"
	"testing"

	"github.com/appcanary/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	conf, err := NewConfFromEnv()
	assert.Nil(err)

	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	assert.Equal(4, len(conf.Watchers), "len of files")

	dpkg := conf.Watchers[0]
	assert.Equal("/var/lib/dpkg/available", dpkg.Path, "file path")

	gemfile := conf.Watchers[1]
	assert.Equal("/path/to/Gemfile.lock", gemfile.Path, "file path")

	tar_h := conf.Watchers[2]
	assert.Equal("fakecmdhere", tar_h.Command, "command path")

	inspectProcess := conf.Watchers[3]
	assert.Equal("*", inspectProcess.Process, "inspect process pattern")

	assert.Equal("123456", conf.ServerConf.UUID)
}

func TestConfUpgrade(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	// everything is configured to default, but file is missing
	assert.Nil(os.Rename(DEV_CONF_FILE, DEV_CONF_FILE+".bak"))
	assert.Nil(os.Rename(DEV_VAR_FILE, DEV_VAR_FILE+".bak"))
	assert.False(fileExists(DEV_CONF_FILE))

	// ensure we have the old dev conf file it'll fall back on,
	// convert and rename
	assert.True(fileExists(OLD_DEV_CONF_FILE))

	// now do the conversion
	conf, err := NewConfFromEnv()
	assert.Nil(err)

	// check that the configuration is ok
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	// ensure old ones got copied to .deprecated and new yaml files exist
	assert.False(fileExists(OLD_DEV_CONF_FILE))
	assert.True(fileExists(OLD_DEV_CONF_FILE + ".deprecated"))
	assert.True(fileExists(OLD_DEV_VAR_FILE + ".deprecated"))

	assert.True(fileExists(DEV_CONF_FILE))

	// great. Now let's ensure this new file is readable.
	// let's make sure we're not reading the old file
	assert.False(fileExists(OLD_DEV_CONF_FILE))

	conf, err = NewConfFromEnv()
	assert.Nil(err)
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	// now we clean up the renamed test files

	assert.Nil(os.Rename(DEV_CONF_FILE+".bak", DEV_CONF_FILE))
	assert.Nil(os.Rename(DEV_VAR_FILE+".bak", DEV_VAR_FILE))

	assert.Nil(os.Rename(OLD_DEV_CONF_FILE+".deprecated", OLD_DEV_CONF_FILE))
	assert.Nil(os.Rename(OLD_DEV_VAR_FILE+".deprecated", OLD_DEV_VAR_FILE))
}

func TestCustomConfPathTOMLConf(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	// link files to a non-standard location
	assert.Nil(os.Link(OLD_DEV_CONF_FILE, "/tmp/agent.conf"))
	assert.Nil(os.Link(OLD_DEV_VAR_FILE, "/tmp/server.conf"))

	// set the new values in the environment
	env.ConfFile = "/tmp/agent.conf"
	env.VarFile = "/tmp/server.conf"

	// attempt to load the config
	conf, err := NewConfFromEnv()

	// there should be an error
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "Is this file valid YAML?"))

	// there should not be a configuration
	assert.Nil(conf)

	// ditch the links
	assert.Nil(os.Remove("/tmp/agent.conf"))
	assert.Nil(os.Remove("/tmp/server.conf"))
}

func TestCustomConfPathYAMLConf(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	// link files to a non-standard location
	assert.Nil(os.Link(DEV_CONF_FILE, "/tmp/agent.yml"))
	assert.Nil(os.Link(DEV_VAR_FILE, "/tmp/server.yml"))

	// set the new values in the environment
	env.ConfFile = "/tmp/agent.yml"
	env.VarFile = "/tmp/server.yml"

	// attempt to load the configuration
	conf, err := NewConfFromEnv()

	// there should not be an error
	assert.Nil(err)

	// there SHOULD be a conf with some things in it
	assert.NotNil(conf)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)
	assert.Equal("123456", conf.ServerConf.UUID)

	// ditch the links
	assert.Nil(os.Remove("/tmp/agent.yml"))
	assert.Nil(os.Remove("/tmp/server.yml"))
}
