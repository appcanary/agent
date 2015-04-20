package agent

import (
	. "github.com/stateio/canary-agent/agent/models"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var log = umwelten.Log

type Agent struct {
	conf   *Conf
	client Client
	server *Server
	files  WatchedFiles
}

func NewAgent(conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, files: WatchedFiles{}}

	if len(clients) > 0 {
		agent.client = clients[0]
	} else {
		agent.client = NewClient(conf.ApiKey, conf.ServerName)
	}

	// what do we know about this machine?
	agent.server = ThisServer(conf.Server.UUID)

	// start watching files
	for _, f := range conf.Files {
		agent.files = append(agent.files, NewWatchedFile(f.Path))
	}

	// First time ever we boot up on this machine

	return agent
}

// TODO modify to read files.

func (self *Agent) Heartbeat() error {

	return self.client.HeartBeat(self.server.UUID, self.files)
}

func (a *Agent) Submit(name string, data interface{}) {
	err := a.client.Submit(name, data)
	if err != nil {
		log.Fatal(err)
	}
}

func (self *Agent) FirstRun() bool {
	// the configuration didn't find a server uuid
	return self.server.IsNew()
}

func (self *Agent) RegisterServer() error {
	err := self.client.CreateServer(self.server)
	log.Debug("Registered server, got: %s", self.server.UUID)

	if err != nil {
		return err
	}

	self.UpdateConf()
	return nil
}

func (self *Agent) UpdateConf() {
	self.conf.Server.UUID = self.server.UUID

	self.conf.PersistServerConf(env)
}

// This has to be called before exiting
func (a *Agent) CloseWatches() {
	for _, file := range a.files {
		file.RemoveHook()
	}
}
