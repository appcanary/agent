package agent

import (
	. "github.com/stateio/canary-agent/agent/models"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var log = umwelten.Log

type Agent struct {
	conf   *Conf
	apps   map[string]*App
	client Client
	server *Server
	files  WatchedFiles
}

func NewAgent(conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, apps: map[string]*App{}, files: WatchedFiles{}}

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

	app_slice := make([]*App, len(self.apps), len(self.apps))

	i := 0
	for _, val := range self.apps {
		app_slice[i] = val
		i = i + 1
	}

	return self.client.HeartBeat(self.server.UUID, app_slice)
}

func (a *Agent) Submit(name string, data interface{}) {
	err := a.client.Submit(name, data)
	if err != nil {
		log.Fatal(err)
	}
}

// func (self *Agent) AddApp(name string, filepath string, appType AppType) *App {
// if self.apps[name] != nil {
// 	log.Fatal("Already have an app ", name)
// }

// application := &App{Name: name, Path: filepath, MonitoredFiles: filepath, AppType: appType, Callback: self.Submit}
// self.apps[name] = application

// if appType == RubyApp {
// 	f := &Gemfile{Path: path.Join(filepath, "Gemfile.lock")}
// 	application.WatchFile(f)
// } else {
// 	log.Fatal("Unrecognized app type ", appType)
// }
// return application
// }

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

// func (self *Agent) RegisterApps() (err error) {
// 	for _, app := range self.apps {
// 		// don't register apps that have been registered
// 		if app.IsNew() {
// 			app.UUID, err = self.client.CreateApp(self.server.UUID, app)
// 			if err != nil {
// 				return err
// 			}
//
// 			log.Debug("Registered app %s, got: %s", app.Name, app.UUID)
// 		}
// 	}
// 	return nil
// }

// This has to be called before exiting
func (a *Agent) CloseWatches() {
	for _, file := range a.files {
		file.RemoveHook()
	}
}
