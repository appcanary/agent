package models

import (
	"io/ioutil"
	"time"

	"github.com/stateio/canary-agent/agent/umwelten"
	"gopkg.in/fsnotify.v1"
)

var log = umwelten.Log

type FileChangeHandler func(*WatchedFile)

type WatchedFile struct {
	Kind         string            `json:"kind"`
	Path         string            `json:"path"`
	UpdatedAt    time.Time         `json:"updated-at"`
	BeingWatched bool              `json:"being-watched"`
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

func (wf *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(wf.Path)
}

func (wf *WatchedFile) RemoveHook() {
	log.Debug("closing watcher")
	wf.Watcher.Close()
}

func (wf *WatchedFile) AddHook() {
	wf.BeingWatched = false
	for !wf.BeingWatched {
		log.Debug("Adding file watcher to %s", wf.Path)
		err := wf.Watcher.Add(wf.Path)
		if err == nil {
			log.Debug("Reading file: %s", wf.Path)
			go wf.OnFileChange(wf)
			wf.BeingWatched = true
		} else {
			log.Error("Failed to add watcher on %s", wf.Path)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (wf *WatchedFile) StartListener() {
	wf.AddHook()

	// Listen for changes
	go func() {
		keepListening := true
		for keepListening {
			select {
			case event, ok := <-wf.Watcher.Events:
				if ok {

					log.Debug("Watcher got event %s", event.String())
					//If the file is renamed or removed we have to create a new watch after a delay
					if isOp(event.Op, fsnotify.Remove) || isOp(event.Op, fsnotify.Rename) {
						//File is deleted so we have to add the watcher again
						go wf.AddHook()
					} else if isOp(event.Op, fsnotify.Write) {
						log.Debug("Rereading file: %s", wf.Path)
						go wf.OnFileChange(wf)
					}
					// else: the op was chmod, do nothing

				} else {
					log.Debug("Closing listener")
					keepListening = false
				}

			case err, ok := <-wf.Watcher.Errors:
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
