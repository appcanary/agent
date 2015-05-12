package agent

import (
	"testing"

	"github.com/stateio/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)

	env.ConfFile = "../test/data/test.conf"
	env.VarFile = "../test/data/test_server.conf"
	conf := NewConfFromEnv()
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal(1, len(conf.Files), "len of files")
	file := conf.Files[0]
	assert.Equal("/foo/bar/baz", file.Path, "file path")

	assert.Equal("123456", conf.Server.UUID)
}
