package agent

import (
	"testing"

	"github.com/appcanary/testify/assert"
	"path/filepath"
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

	assert.Equal("123456", conf.ServerConf.UUID)
}

func TestServerConf(t *testing.T) {
	assert := assert.New(t)
	conf := &ServerConf{}
	file, _ := filepath.Abs("../test/data/issue")
	env.DebianLikeDistributionFile = file
	conf.ParseDistro()
	assert.Equal("Debian GNU/Linux 8 \\n \\l\n\n", conf.DistroString, "parses distro")
}

func TestServerConfUbuntu(t *testing.T) {
	assert := assert.New(t)
	conf := &ServerConf{}
	file, _ := filepath.Abs("../test/data/ubuntu_issue")
	env.DebianLikeDistributionFile = file
	conf.ParseDistro()
	assert.Equal("Ubuntu 14.04.2 LTS \\n \\l\n", conf.DistroString, "parses distro")
}
