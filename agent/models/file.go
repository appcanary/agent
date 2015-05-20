package models

import (
	"io/ioutil"
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
	file.StartListener()
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
	needHook := true
	for needHook {
		log.Debug("Adding file watcher to %s", self.Path)
		err := self.Watcher.Add(self.Path)
		if err == nil {
			log.Debug("Reading file: %s", self.Path)
			go self.OnFileChange(self)
			needHook = false
		} else {
			log.Debug("Failed to add watcher on %s", self.Path)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (self *WatchedFile) StartListener() {
	self.AddHook()

	// Listen for changes
	go func() {
		keepListening := true
		for keepListening {
			select {
			case event, ok := <-self.Watcher.Events:
				if ok {

					log.Debug("Watcher got event %s", event.String())
					//If the file is renamed or removed we have to create a new watch after a delay
					if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {
						//File is deleted so we have to add the watcher again
						go self.AddHook()
					} else if isOp(event.Op, fsnotify.Write) {
						log.Debug("Rereading file: %s", self.Path)
						go self.OnFileChange(self)
					}
					// else: the op was chmod, do nothing

				} else {
					log.Debug("Closing listener")
					keepListening = false
				}

			case err, ok := <-self.Watcher.Errors:
				if ok {
					log.Debug("Watcher error: %s", err)
				} else {
					break
				}
			}
		}
	}()
}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
