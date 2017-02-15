package conf

import (
	"os"
	"testing"

	"github.com/appcanary/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	// origConfFile := "../test/data/test.conf"
	// origVarFile := "../test/data/test_server.conf"

	// env.ConfFile = origConfFile
	// env.VarFile = origVarFile
	conf := NewConfFromEnv()

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

	// rename the test files back again
	// assert.Nil(os.Rename(origConfFile+".deprecated", origConfFile))
	// assert.Nil(os.Rename(origVarFile+".deprecated", origVarFile))
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
	conf := NewConfFromEnv()

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

	conf = NewConfFromEnv()
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
