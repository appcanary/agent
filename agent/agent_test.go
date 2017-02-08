package agent

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/appcanary/agent/agent/conf"
	"github.com/appcanary/testify/assert"
)

func TestAgent(t *testing.T) {
	assert := assert.New(t)

	// setup
	serverUUID := "123456"
	conf.InitEnv("test")
	config := conf.NewTomlConfFromEnv()

	config.Watchers[0].Path = conf.DEV_CONF_PATH + "/dpkg/available"

	client := &MockClient{}
	client.On("CreateServer").Return(serverUUID)
	client.On("SendFile").Return(nil).Twice()
	client.On("Heartbeat").Return(nil).Once()
	client.On("SendProcessState").Return(nil).Twice()

	agent := NewAgent("test", config, client)

	// let's make sure stuff got set
	assert.Equal("deployment1", agent.server.Name)
	assert.NotEqual("", agent.server.Hostname)
	assert.NotEqual("", agent.server.Uname)
	assert.NotEqual("", agent.server.Distro)
	assert.NotEqual("", agent.server.Ip)

	// let's ensure our server is unregistered
	agent.server.UUID = ""

	assert.Equal(true, agent.FirstRun())

	agent.RegisterServer()

	// registering the server actually set the right val
	assert.Equal(serverUUID, agent.server.UUID)

	// Let's ensure that the client gets exercised.
	agent.BuildAndSyncWatchers()
	agent.StartPolling()

	// force a change in the process table
	proc := startProcess(assert)
	defer proc.Kill()

	agent.Heartbeat()

	// after a period of time, we sync all files
	agent.SyncAllFiles()

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

func startProcess(assert *assert.Assertions) *os.Process {
	script := conf.DEV_CONF_PATH + "/pointless"

	cmd := exec.Command(script)
	err := cmd.Start()
	assert.Nil(err)

	return cmd.Process
}
