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
	Kind         string            `json:"kind"`
	Path         string            `json:"path"`
	UpdatedAt    time.Time         `json:"updated-at"`
	Watcher      *fsnotify.Watcher `json:"-"`
	OnFileChange FileChangeHandler `json:"-"`
}

type WatchedFiles []*WatchedFile

// TODO: time.Now() needs to be called whenever it updates
func NewWatchedFileWithHook(path string, callback FileChangeHandler) *WatchedFile {
	file := NewWatchedFile(path, callback)
	file.AddHook()
	return file
}

func NewWatchedFile(path string, callback FileChangeHandler) *WatchedFile {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err.Error())
	}

	file := &WatchedFile{Path: path, OnFileChange: callback, Kind: "gemfile", UpdatedAt: time.Now(), Watcher: watcher}
	return file
}

func (self *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(self.Path)
}

func (self *WatchedFile) RemoveHook() {
	log.Debug("closing watcher")
	self.Watcher.Close()
}

func (self *WatchedFile) AddHook() {
	log.Info("Reading file: %s", self.Path)
	go self.OnFileChange(self)

	log.Info("Starting watcher on %s", self.Path)
	err := self.Watcher.Add(self.Path)

	if err != nil {
		log.Fatal(err.Error())
	}

	go self.ChangeListener()

}

func (self *WatchedFile) ChangeListener() {
	keepListening := true
	for keepListening {
		// shouldBreak := false
		select {
		case event, ok := <-self.Watcher.Events:
			if ok {

				log.Info("Got event %s", event.String())

				//If the file is renamed or removed we have to create a new watch after a delay
				if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {
					go self.HandleRemoved()

				} else if isOp(event.Op, fsnotify.Write) {
					log.Info("Rereading file: %s", self.Path)
					go self.OnFileChange(self)
				}
				// else: the op was chmod, do nothing

			} else {
				keepListening = false
			}

		case err, ok := <-self.Watcher.Errors:
			if ok {
				log.Info("error:", err)
			} else {
				break
			}
		}
	}
}

func (self *WatchedFile) HandleRemoved() {
	log.Info("File moved: %s", self.Path)

	//TODO: be smarter about this delay
	time.Sleep(100 * time.Millisecond)

	// File doesn't exist
	if _, err := os.Stat(self.Path); err != nil {
		// TODO: this is something we should handle gracefully with a expanding timeout, and an error sent to our server
		log.Fatal(err)
	}

	err := self.Watcher.Add(self.Path)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Rereading file after move: %s", self.Path)
	go self.OnFileChange(self)

}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
