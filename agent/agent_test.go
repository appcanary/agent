package agent

import (
	"testing"
	"time"

	"github.com/appcanary/agent/agent/umwelten"
	"github.com/appcanary/agent/mocks"
	"github.com/appcanary/testify/assert"
)

func TestAgent(t *testing.T) {
	assert := assert.New(t)

	// setup
	umwelten.Init("test")
	conf := NewConfFromEnv()

	// conf paths are absolute, which will
	// fail across diff testing envs.
	conf.Files[0].Path = env.ConfFile

	client := &mocks.Client{}
	agent := NewAgent(conf, client)

	// let's ensure our server is unregistered
	agent.server.UUID = ""

	assert.Equal(true, agent.FirstRun())
	server_uuid := "123456"

	client.On("CreateServer").Return(server_uuid)
	agent.RegisterServer()

	// registering the server actually set the right val
	assert.Equal(server_uuid, agent.server.UUID)

	// Let's ensure that the client gets exercised.
	client.On("SendFile").Return(nil).Once()
	agent.StartWatching()

	client.On("Heartbeat").Return(nil).Once()
	agent.Heartbeat()

	// the filewatcher needs enough time to
	// actually be able to start watching
	// the file. This is clunky, but less clunky
	// than hacking some channel into this.
	<-time.After(200 * time.Millisecond)
	// close the hooks before asserting expectations
	// since the SendFiles happen in a go routine
	defer agent.CloseWatches()
	defer client.AssertExpectations(t)
}
