package models

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/stateio/canary-agent/agent/umwelten"
	"gopkg.in/fsnotify.v1"
)

var log = umwelten.Log

type File interface {
	GetPath() string
	Parse() interface{}
}

type FileChangeHandler func(*WatchedFile)

type WatchedFile struct {
	Name         string            `json:"name"`
	Path         string            `json:"monitoredFiles"`
	Watcher      *fsnotify.Watcher `json:"-"`
	OnFileChange FileChangeHandler `json:"-"`
}

type WatchedFiles []*WatchedFile

func NewWatchedFile(path string, callback FileChangeHandler) *WatchedFile {
	file := &WatchedFile{Path: path, OnFileChange: callback}
	file.AddHook()
	return file
}

func (self *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(self.Path)
}

// TODO: make this a finalizer? :(
func (self *WatchedFile) RemoveHook() {
	log.Info("closing watcher")
	self.Watcher.Close()
}

func (self *WatchedFile) AddHook() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting watcher on %s", self.Path)
	self.Watcher = watcher

	go func() {
		for {
			select {
			case event, more := <-watcher.Events:
				if more {

					log.Info("Got event %s", event.String())

					//If the file is renamed or removed we have to create a new watch after a delay
					if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {

						go func() {
							log.Info("File moved: %s", self.Path)

							//TODO: be smarter about this delay
							time.Sleep(100 * time.Millisecond)

							// File doesn't exist
							if _, err := os.Stat(self.Path); err != nil {
								// TODO: this is something we should handle gracefully with a expanding timeout, and an error sent to our server
								log.Fatal(err)
							}

							err = self.Watcher.Add(self.Path)

							if err != nil {
								log.Fatal(err)
							}

							log.Info("Rereading file after move: %s", self.Path)
							go self.OnFileChange(self)
						}()

					} else if isOp(event.Op, fsnotify.Write) {
						log.Info("Rereading file: %s", self.Path)
						go self.OnFileChange(self)
						// TODO commented out for now
						// go a.Submit(wf.Parse())
					} // else: the op was chmod, do nothing

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

	log.Info("Reading file: %s", self.Path)

	go self.OnFileChange(self)

	err = self.Watcher.Add(self.Path)
	if err != nil {
		log.Fatal(err.Error())
	}

}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
