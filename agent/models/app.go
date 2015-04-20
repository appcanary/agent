package models

import "github.com/stateio/canary-agent/agent/umwelten"

var log = umwelten.Log

type Submitter func(string, interface{})

type App struct {
	Name           string  `json:"name"`
	Path           string  `json:"-"`
	AppType        AppType `json:"-"`
	watchedFiles   WatchedFiles
	MonitoredFiles string    `json:"monitoredFiles"`
	Callback       Submitter `json:"-"`
	UUID           string    `json:"-"`
}

type AppType int

const (
	UnknownApp AppType = iota
	RubyApp
)

func (self *App) IsNew() bool {
	return self.UUID == ""
}

func (a *App) Submit(data interface{}) {
	a.Callback(a.Name, data)
}
