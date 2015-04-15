package agent

import (
	"path"

	"github.com/stateio/canary-agent/agent/app"
	"github.com/stateio/canary-agent/agent/server"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var log = umwelten.Log

type Agent struct {
	conf   *Conf
	apps   map[string]*app.App
	client Client
	server *server.Server
}

func NewAgent(conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, apps: map[string]*app.App{}}

	if len(clients) > 0 {
		agent.client = clients[0]
	} else {
		agent.client = NewClient(conf.ApiKey, conf.ServerName)
	}

	// what do we know about thi machine?
	agent.server = server.New()

	// COMMENTED OUT FOR NOW
	// load the existing gemfiles
	// for _, a := range conf.Apps {
	// 	if a.Type == "ruby" {
	// 		agent.AddApp(a.Name, a.Path, RubyApp)
	// 	}
	// }

	// First time ever we boot up on this machine

	return agent
}

func (self *Agent) Heartbeat() error {
	return self.client.HeartBeat(self.server.UUID)
}

func (a *Agent) Submit(name string, data interface{}) {
	err := a.client.Submit(name, data)
	if err != nil {
		log.Fatal(err)
	}
}

func (self *Agent) AddApp(name string, filepath string, appType app.AppType) *app.App {
	if a.apps[name] != nil {
		log.Fatal("Already have an app ", name)
	}

	application := &app.App{Name: name, Path: filepath, AppType: appType, Callback: self.Submit}
	self.apps[name] = application

	if appType == app.RubyApp {
		f := &Gemfile{Path: path.Join(filepath, "Gemfile.lock")}
		application.WatchFile(f)
	} else {
		log.Fatal("Unrecognized app type ", appType)
	}
	return application
}

func (self *Agent) RegisterServer() error {
	err := self.client.CreateServer(self.server)

	log.Debug("Registered server, got: " + self.server.UUID)

	if err != nil {
		return err
	}
	return nil
}

// This has to be called before exiting
func (a *Agent) CloseWatches() {
	for _, appli := range a.apps {
		appli.CloseWatches()
	}
}
