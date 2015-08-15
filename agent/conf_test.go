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
	env.DistributionFile = file
	conf.ParseDistro()
	assert.Equal("debian", conf.Distro, "parses distro")
	assert.Equal("8", conf.Release, "parses release")
}

func TestServerConfUbuntu(t *testing.T) {
	assert := assert.New(t)
	conf := &ServerConf{}
	file, _ := filepath.Abs("../test/data/ubuntu_issue")
	env.DistributionFile = file
	conf.ParseDistro()
	assert.Equal("ubuntu", conf.Distro, "parses distro")
	assert.Equal("14.04.2", conf.Release, "parses release")
}
