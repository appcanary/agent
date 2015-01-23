package agent

import (
	"testing"
	"time"

	"github.com/stateio/canary-agent/mocks"
	"github.com/stateio/canary-agent/parsers/gemfile"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	assert := assert.New(t)
	conf := NewConf()
	conf.ServerName = "test-server"
	conf.Apps = []AppConf{AppConf{Name: "test", Type: "ruby", Path: "./testdata/"}}
	client := &mocks.Client{}
	//Make sure that we call a heartbeat and register the server
	client.On("HeartBeat").Return(nil).Once()
	gemfile, _ := gemfile.ParseGemfile("testdata/Gemfile.lock")
	_ = gemfile
	client.On("Submit", "test", gemfile).Return(nil).Once()

	agent := NewAgent(conf, client)
	defer agent.CloseWatches()
	assert.Equal(1, len(agent.apps), "len agent.apps")
	app := agent.apps["test"]

	assert.Equal("test", app.Name, "app name")
	assert.Equal(RubyApp, app.AppType, "app type")
	assert.Equal("./testdata/", app.Path, "app path")
	assert.Equal(1, len(app.watchedFiles), "len app.WatchedFiles")

	wf := app.watchedFiles[0]
	assert.Equal("testdata/Gemfile.lock", wf.GetPath(), "gem file path")
	//some of the submits are called in goroutines so we need to wait a bit for them to finish
	time.Sleep(100 * time.Millisecond)
	client.AssertExpectations(t)
}
