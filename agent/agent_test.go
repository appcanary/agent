package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	assert := assert.New(t)
	conf := NewConf()
	conf.ServerName = "test"
	conf.Apps = []AppConf{AppConf{Name: "test", Type: "ruby", Path: "./testdata/"}}
	agent := NewAgent(conf)
	defer agent.CloseWatches()

	assert.Equal(1, len(agent.apps), "len agent.apps")
	app := agent.apps["test"]

	assert.Equal("test", app.Name, "app name")
	assert.Equal(RubyApp, app.AppType, "app type")
	assert.Equal("./testdata/", app.Path, "app path")
	assert.Equal(1, len(app.WatchedFiles), "len app.WatchedFiles")

	wf := app.WatchedFiles[0]
	assert.Equal("testdata/Gemfile.lock", wf.GetPath(), "gem file path")

}
