package models

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/stateio/canary-agent/agent/umwelten"
	"gopkg.in/fsnotify.v1"
)

var log = umwelten.Log

type FileChangeHandler func(*WatchedFile)

type WatchedFile struct {
	lock         sync.RWMutex
	keepPolling  bool
	Kind         string            `json:"kind"`
	Path         string            `json:"path"`
	UpdatedAt    time.Time         `json:"updated-at"`
	BeingWatched bool              `json:"being-watched"`
	Watcher      *fsnotify.Watcher `json:"-"`
	OnFileChange FileChangeHandler `json:"-"`
	state        *os.FileInfo      `json:"-"`
}

type WatchedFileJson struct {
	Kind         string    `json:"kind"`
	Path         string    `json:"path"`
	UpdatedAt    time.Time `json:"updated-at"`
	BeingWatched bool      `json:"being-watched"`
}

type WatchedFiles []*WatchedFile

// TODO: time.Now() needs to be called whenever it updates
func NewWatchedFileWithHook(path string, callback FileChangeHandler) *WatchedFile {
	file := NewWatchedFile(path, callback)
	file.StartListener()
	return file
}

func NewWatchedFile(path string, callback FileChangeHandler) *WatchedFile {
	file := &WatchedFile{Path: path, OnFileChange: callback, Kind: "gemfile", UpdatedAt: time.Now()}
	return file
}

func (wf *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(wf.Path)
}

// Rename to StopListener
// TODO: solve data race issue
func (wf *WatchedFile) RemoveHook() {
	log.Debug("closing watcher")
	wf.keepPolling = false
}

func (wf *WatchedFile) GetBeingWatched() bool {
	wf.lock.RLock()
	defer wf.lock.RUnlock()
	return wf.BeingWatched
}

func (wf *WatchedFile) SetBeingWatched(bw bool) {
	wf.lock.Lock()
	wf.BeingWatched = bw
	wf.lock.Unlock()
}

// func (wf *WatchedFile) MarshalJson() ([]byte, error) {
// 	wf.lock.RLock()
// 	defer wf.lock.RUnlock()
// 	return json.Marshal(interface{}(wf))
// }

func (wf *WatchedFile) listen() {
	for wf.keepPolling {
		// TBD do we want to stop this EVER? prob no
		// if !wf.GetBeingWatched() {
		// 	return
		// }

		wf.scan()
		time.Sleep(250 * time.Millisecond)
	}
}

func (wf *WatchedFile) scan() {
	// log.Debug("SCANNING...")
	info, err := os.Stat(wf.Path)

	// log.Debug("inf", info)
	if err != nil {
		wf.SetBeingWatched(false)
		log.Debug("File Stat error")
		return // try again later?
	}

	wf.SetBeingWatched(true)

	if wf.fileChanged(wf.state, &info) {
		// log.Debug("FILE CHANGE")
		go wf.OnFileChange(wf)
		wf.state = &info
	}
}

func (wf *WatchedFile) fileChanged(fptr1 *os.FileInfo, fptr2 *os.FileInfo) bool {
	if fptr1 == nil {
		return true
	}

	if fptr2 == nil {
		return true // TBD
	}

	file1 := *fptr1
	file2 := *fptr2

	// fmt.Printf("f1 %+v\n", file1)
	// fmt.Printf("f2 %+v\n\n", file2)

	return file1.Size() != file2.Size() || file1.ModTime() != file2.ModTime() || file1.Mode() != file2.Mode()
}

func (wf *WatchedFile) StartListener() {
	wf.keepPolling = true
	go wf.listen()
}

// Checks whether an fsnotify Op from an event matches a target Op
func isOp(o fsnotify.Op, target fsnotify.Op) bool {
	return (o&target == target)
}
