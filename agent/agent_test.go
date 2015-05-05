package agent

import (
	"testing"

	"github.com/stateio/canary-agent/agent/umwelten"
	"github.com/stateio/canary-agent/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	assert := assert.New(t)

	// setup
	umwelten.Init("test")
	conf := NewConfFromEnv()
	client := &mocks.Client{}
	agent := NewAgent(conf, client)

	// let's ensure our server is unregistered
	agent.server.UUID = ""

	assert.Equal(agent.FirstRun(), true)
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

	// close the hooks before asserting expectations
	// since the SendFiles happen in a go routine
	agent.CloseWatches()
	client.AssertExpectations(t)
}
