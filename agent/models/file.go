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
	Kind          string            `json:"kind"`
	Path          string            `json:"path"`
	UpdatedAt     time.Time         `json:"updated-at"`
	Watcher       *fsnotify.Watcher `json:"-"`
	OnFileChange  FileChangeHandler `json:"-"`
	resetListener chan bool
	done          chan bool
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

	file := &WatchedFile{Path: path, OnFileChange: callback, Kind: "gemfile", UpdatedAt: time.Now(), Watcher: watcher, resetListener: make(chan bool, 1), done: make(chan bool, 1)}
	return file
}

func (self *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(self.Path)
}

func (self *WatchedFile) RemoveHook() {
	log.Debug("closing watcher")
	self.done <- true
	self.Watcher.Close()
}

func (self *WatchedFile) AddHook() {
	f, _ := os.OpenFile("/tmp/wtf", os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	f.WriteString("Addhook1\n")
	self.resetListener <- true
	go self.ChangeListener()

	go func() {
		f, _ := os.OpenFile("/tmp/wtf", os.O_APPEND|os.O_WRONLY, 0600)
		defer f.Close()
		f.WriteString("begin hook loop\n")
		for {
			select {
			case <-self.done:
				return
			case <-self.resetListener:
				// keep trying to listen to this, in perpetuity.
				log.Debug("Adding file watcher to %s", self.Path)
				f.WriteString("adding file watcher\n")
				err := self.Watcher.Add(self.Path)

				if err == nil {
					log.Debug("Reading file: %s", self.Path)
					f.WriteString("read file\n")
					go self.OnFileChange(self)

				} else {
					log.Debug("Failed to add watcher on %s", self.Path)

					f.WriteString("failed to add\n")

					// try again in a bit, arbitrary time limit
					go func() {
						// sleep := time.After(100 * time.Millisecond)
						// <-sleep
						self.resetListener <- true
					}()
				}
			}
		}
	}()

}

func (self *WatchedFile) ChangeListener() {
	f, _ := os.OpenFile("/tmp/wtf", os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	keepListening := true
	for keepListening {
		select {
		case event, ok := <-self.Watcher.Events:
			if ok {

				log.Debug("Watcher got event %s", event.String())

				//If the file is renamed or removed we have to create a new watch after a delay
				if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {
					f.WriteString("resetlistener\n")
					self.resetListener <- true

				} else if isOp(event.Op, fsnotify.Write) {
					log.Debug("Rereading file: %s", self.Path)
					f.WriteString("rereading file\n")
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
				f.WriteString("watch errar\n")
			} else {
				break
			}
		}
	}
}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
