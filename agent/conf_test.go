package agent

import (
	"testing"

	"github.com/appcanary/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)

	env.ConfFile = "../test/data/test.conf"
	env.VarFile = "../test/data/test_server.conf"
	conf := NewConfFromEnv()
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal(2, len(conf.Files), "len of files")
	dpkg := conf.Files[0]
	assert.Equal("/var/lib/dpkg/available", dpkg.Path, "file path")

	gemfile := conf.Files[1]
	assert.Equal("/path/to/Gemfile.lock", gemfile.Path, "file path")

	assert.Equal("123456", conf.Server.UUID)
}
