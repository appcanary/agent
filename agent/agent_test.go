package agent

import (
	"testing"
	"time"

	"github.com/appcanary/testify/assert"
)

func TestAgent(t *testing.T) {
	assert := assert.New(t)

	// setup
	InitEnv("test")
	conf := NewConfFromEnv()

	conf.Files[0].Path = DEV_CONF_PATH + "/dpkg/available"

	client := &MockClient{}
	agent := NewAgent("test", conf, client)

	// let's make sure stuff got set
	assert.Equal("deployment1", agent.server.Name)
	assert.NotEqual("", agent.server.Hostname)
	assert.NotEqual("", agent.server.Uname)
	assert.NotEqual("", agent.server.Distro)
	assert.NotEqual("", agent.server.Ip)

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
