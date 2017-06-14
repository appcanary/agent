package agent

import (
	"testing"

	"github.com/appcanary/agent/agent/detect"
	"github.com/appcanary/agent/conf"
	"github.com/appcanary/testify/assert"
)

func TestServerConf(t *testing.T) {
	assert := assert.New(t)

	aconf := &conf.Conf{
		LinuxOSInfo: detect.LinuxOSInfo{
			Distro:  "testDistro",
			Release: "testRelease",
		},
		ServerName: "TestName",
		Tags: []string{
			"simon",
			"dogs",
		},
	}
	server := NewServer(aconf, &conf.ServerConf{})

	assert.Equal("testDistro", server.Distro)
	assert.Equal("testRelease", server.Release)
	assert.Equal("TestName", server.Name)
	assert.Equal([]string{"simon", "dogs"}, server.Tags)

	aconf = &conf.Conf{}
	server = NewServer(aconf, &conf.ServerConf{})

	// amusingly, can't test generated values reliably
	// because these tests run in unpredictable linuxes
	assert.NotEqual("testDistro", server.Distro)
	assert.NotEqual("testRelease", server.Release)
	assert.Equal("", server.Name)

}
