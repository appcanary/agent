package agent

import (
	"fmt"
	"path"

	"github.com/op/go-logging"
	"github.com/stateio/canary-agent/agent/server"
)

var lg = logging.MustGetLogger("app-canary")

type Agent struct {
	conf   *Conf
	apps   map[string]*App
	client Client
	server *server.Server
}

type App struct {
	Name         string
	Path         string
	AppType      AppType
	watchedFiles WatchedFiles
	callback     Submitter
}

type AppType int

const (
	UnknownApp AppType = iota
	RubyApp
)

type Submitter func(string, interface{})

func NewAgent(conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, apps: map[string]*App{}}
	if len(clients) > 0 {
		agent.client = clients[0]
	} else {
		agent.client = NewClient(conf.ApiKey, conf.ServerName)
	}

	// what do we know about thi machine?
	agent.server = server.New()

	err := agent.RegisterServer()
	lg.Debug("Registered server, got: " + agent.server.UUID)
	if err != nil {
		lg.Fatal(err)
	}

	// COMMENTED OUT FOR NOW
	// load the existing gemfiles
	// for _, a := range conf.Apps {
	// 	if a.Type == "ruby" {
	// 		agent.AddApp(a.Name, a.Path, RubyApp)
	// 	}
	// }

	// First time ever we boot up on this machine

	err = agent.client.HeartBeat()
	if err != nil {
		lg.Fatal(err)
	}

	return agent
}

func (a *Agent) Submit(name string, data interface{}) {
	err := a.client.Submit(name, data)
	if err != nil {
		lg.Fatal(err)
	}
}

func (a *Agent) AddApp(name string, filepath string, appType AppType) *App {
	if a.apps[name] != nil {
		panic(fmt.Sprintf("Already have an app %s", name))
	}

	app := &App{Name: name, Path: filepath, AppType: appType, callback: a.Submit}
	a.apps[name] = app

	if appType == RubyApp {
		f := &Gemfile{Path: path.Join(filepath, "Gemfile.lock")}
		app.WatchFile(f)
	} else {
		panic(fmt.Sprintf("Unrecognized app type %s", appType))
	}
	return app
}

func (a *Agent) RegisterServer() error {
	return a.client.CreateServer(a.server)
}

// This has to be called before exiting
func (a *Agent) CloseWatches() {
	for _, app := range a.apps {
		app.CloseWatches()
	}
}
