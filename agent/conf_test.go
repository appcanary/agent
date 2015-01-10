package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)

	conf := NewConfFromFile("testdata/test.conf")
	assert.Equal("deployment1", conf.ServerName, "server_name")
	assert.Equal("APIKEY", conf.ApiKey)
	assert.Equal(true, conf.TrackSystemPackages, "track_system_packages")
	assert.Equal("info", conf.LogLevel, "log_level")
	assert.Equal(1, len(conf.Apps), "len of apps")
	app := conf.Apps[0]
	assert.Equal("my cool app", app.Name, "app name")
	assert.Equal("ruby", app.Type, "app type")
	assert.Equal("/foo/bar/baz", app.Path, "app path")
}
