package models

import (
	"os"
	"time"

	"github.com/stateio/canary-agent/agent/umwelten"
	"gopkg.in/fsnotify.v1"
)

var log = umwelten.Log

type Submitter func(string, interface{})

type App struct {
	Name           string  `json:"name"`
	Path           string  `json:"-"`
	AppType        AppType `json:"-"`
	watchedFiles   WatchedFiles
	MonitoredFiles string    `json:"monitoredFiles"`
	Callback       Submitter `json:"-"`
}

type AppType int

const (
	UnknownApp AppType = iota
	RubyApp
)

func (a *App) Submit(data interface{}) {
	a.Callback(a.Name, data)
}

func (a *App) WatchFile(f File) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting watcher on %s", f.GetPath())
	wf := &WatchedFile{File: f, Watcher: watcher}

	go func() {
		for {
			select {
			case event, more := <-watcher.Events:
				if more {

					log.Info("Got event %s", event.String())

					//If the file is renamed or removed we have to create a new watch after a delay
					if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {

						go func() {
							log.Info("File moved: %s", wf.GetPath())

							//TODO: be smarter about this delay
							time.Sleep(100 * time.Millisecond)

							// File doesn't exist
							if _, err := os.Stat(wf.GetPath()); err != nil {
								// TODO: this is something we should handle gracefully with a expanding timeout, and an error sent to our server
								log.Fatal(err)
							}

							err = wf.Watcher.Add(wf.GetPath())

							if err != nil {
								log.Fatal(err)
							}

							log.Info("Rereading file after move: %s", wf.GetPath())
							// TODO commented out for now
							// go a.Submit(wf.Parse())
						}()

					} else if isOp(event.Op, fsnotify.Write) {
						log.Info("Rereading file: %s", wf.GetPath())
						// TODO commented out for now
						// go a.Submit(wf.Parse())
					} // else: the op was chmod, do nothing
					//go a.Submit(wf.Parse())
				} else {
					break //done = true
				}
			case err, more := <-watcher.Errors:
				if more {
					log.Info("error:", err)
				} else {
					break
				}
			}
		}
	}()
	log.Info("Reading file: %s", wf.GetPath())
	// TODO commented out for now
	// go a.Submit(wf.Parse())
	err = wf.Watcher.Add(wf.GetPath())
	if err != nil {
		log.Fatal(err.Error())
	}

	//Add watched file to the apps list
	a.watchedFiles = append(a.watchedFiles, wf)
}

func (a *App) CloseWatches() {
	for _, wf := range a.watchedFiles {
		wf.Close()
	}
}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
