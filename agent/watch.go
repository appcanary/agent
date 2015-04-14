package agent

import (
	"os"
	"time"

	"github.com/stateio/canary-agent/parsers/gemfile"
	"gopkg.in/fsnotify.v1"
)

type File interface {
	GetPath() string
	Parse() interface{}
}

type WatchedFile struct {
	File
	Watcher *fsnotify.Watcher
}

type WatchedFiles []*WatchedFile

type Gemfile struct {
	Path string
}

func (g *Gemfile) GetPath() string {
	return g.Path
}

func (a *App) Submit(data interface{}) {
	a.callback(a.Name, data)
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
					if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {
						//If the file is renamed or removed we have to create a new watch after a delay
						go func() {
							log.Info("File moved: %s", wf.GetPath())
							//TODO: be smarter about this delay
							time.Sleep(100 * time.Millisecond)
							if _, err := os.Stat(wf.GetPath()); err != nil {
								// File doesn't exist
								// TODO: this is something we should handle gracefully with a expanding timeout, and an error sent to our server
								log.Fatal(err)
							}
							err = wf.Watcher.Add(wf.GetPath())
							if err != nil {
								log.Fatal(err)
							}
							log.Info("Rereading file after move: %s", wf.GetPath())
							go a.Submit(wf.Parse())
						}()
					} else if isOp(event.Op, fsnotify.Write) {
						log.Info("Rereading file: %s", wf.GetPath())
						go a.Submit(wf.Parse())
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
	go a.Submit(wf.Parse())
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

func (g *Gemfile) Parse() interface{} {
	gf, err := gemfile.ParseGemfile(g.Path)
	if err != nil {
		//TODO handle error more gracefully
		//If we can't parse try again in a bit
		log.Fatal(err)
	}
	return gf
}

// TODO: make this a finalizer? :(
func (wf *WatchedFile) Close() {
	log.Info("closing watcher")
	wf.Watcher.Close()
}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
