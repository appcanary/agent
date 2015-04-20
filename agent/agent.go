package agent

import (
	"bytes"
	"encoding/base64"

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
		agent.files = append(agent.files, NewWatchedFile(f.Path, agent.OnFileChange))
	}

	// First time ever we boot up on this machine

	return agent
}

func (self *Agent) OnFileChange(file *WatchedFile) {
	log.Info("File change: %s", file.Path)

	// should probably be in the actual hook code
	contents, err := file.Contents()

	if err != nil {
		// again, we need to recover from this
		log.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	b64enc := base64.NewEncoder(base64.StdEncoding, buffer)
	b64enc.Write(contents)
	b64enc.Close()

	// TODO queue this up somehow?
	self.client.SendFile(file.Path, buffer.Bytes())
	// fmt.Printf("\n%s\n", buffer)

}

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
