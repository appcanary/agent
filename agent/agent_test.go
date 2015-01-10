package agent

import (
	"testing"

	"github.com/mveytsman/canary-agent/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	assert := assert.New(t)
	conf := NewConf()
	conf.ServerName = "test"
	conf.Apps = []AppConf{AppConf{Name: "test", Type: "ruby", Path: "./testdata/"}}
	agent := NewAgent(conf, &mocks.Client{})
	defer agent.CloseWatches()

	assert.Equal(1, len(agent.apps), "len agent.apps")
	app := agent.apps["test"]

	assert.Equal("test", app.Name, "app name")
	assert.Equal(RubyApp, app.AppType, "app type")
	assert.Equal("./testdata/", app.Path, "app path")
	assert.Equal(1, len(app.watchedFiles), "len app.WatchedFiles")

	wf := app.watchedFiles[0]
	assert.Equal("testdata/Gemfile.lock", wf.GetPath(), "gem file path")
}

// //func TestAddApp(t *testing.T) {
// //	agent := &Agent{apps: map[string]*App{}}
// 	agent.AddApp("test", "./testdata", RubyApp)
// }
