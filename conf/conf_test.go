package conf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/appcanary/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	origConfFile := "../test/data/test.conf"
	origVarFile := "../test/data/test_server.conf"

	env.ConfFile = origConfFile
	env.VarFile = origVarFile
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

	// rename the test files back again
	assert.Nil(os.Rename(origConfFile+".deprecated", origConfFile))
	assert.Nil(os.Rename(origVarFile+".deprecated", origVarFile))
}

func TestConfUpgrade(t *testing.T) {
	assert := assert.New(t)
	InitEnv("test")

	env.ConfFile = "../test/data/tmptest.conf"
	env.VarFile = "../test/data/tmptest_server.conf"
	// confFile := "../test/data/tmptest.conf"
	// varFile := "../test/data/tmptest_server.conf"

	// set up disposable config
	cp := exec.Command("cp", "../test/data/test.conf", env.ConfFile)
	err := cp.Run()
	assert.Nil(err)

	cp = exec.Command("cp", "../test/data/test_server.conf", env.VarFile)
	err = cp.Run()
	assert.Nil(err)

	// now do the conversion
	conf, err := NewConfFromEnv()
	assert.Nil(err)

	// check the new settings
	newConfFile, err := filepath.Abs("../test/data/tmptest.yml")
	assert.Nil(err)

	newVarFile, err := filepath.Abs("../test/data/tmptest_server.yml")
	assert.Nil(err)

	assert.True(fileExists(newConfFile))
	assert.True(fileExists(newVarFile))

	// check that the configuration is ok
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)

	// now remove the old configs and reload
	rm := exec.Command(
		"rm",
		"../test/data/tmptest.conf.deprecated",
		"../test/data/tmptest_server.conf.deprecated")
	err = rm.Run()
	assert.Nil(err)

	// TODO some mocking so we can check that FullSave isn't called again, or
	// whatever.

	env.ConfFile = newConfFile
	env.VarFile = newVarFile

	conf, err = NewConfFromEnv()
	assert.Nil(err)

	// check that the configuration is ok
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal("deployment1", conf.ServerName)
	assert.Equal("testDistro", conf.Distro)
	assert.Equal("testRelease", conf.Release)
}
