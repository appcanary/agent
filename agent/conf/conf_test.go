package conf

import (
	"testing"

	"github.com/appcanary/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)

	env.ConfFile = "../../test/data/test.conf"
	env.VarFile = "../../test/data/test_server.conf"
	conf := NewTomlConfFromEnv()

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
