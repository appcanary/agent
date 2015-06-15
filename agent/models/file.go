package models

import (
	"io/ioutil"
	"sync"
	"time"

	"hash/crc32"

	"github.com/stateio/canary-agent/agent/umwelten"
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
	OnFileChange FileChangeHandler `json:"-"`
	checksum     uint32
}

type WatchedFiles []*WatchedFile

// TODO: time.Now() needs to be called whenever it updates
func NewWatchedFileWithHook(path string, callback FileChangeHandler) *WatchedFile {
	file := NewWatchedFile(path, callback)
	file.StartListener()
	return file
}

// only used for tests
func NewWatchedFile(path string, callback FileChangeHandler) *WatchedFile {
	file := &WatchedFile{Path: path, OnFileChange: callback, Kind: "gemfile", UpdatedAt: time.Now()}
	return file
}

func (wf *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(wf.Path)
}

// TODO: solve data race issue
func (wf *WatchedFile) StopListening() {
	log.Debug("No longer listening to: ", wf.Path)
	wf.keepPolling = false
}

// TODO: rename
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

		wf.scan()
		// TODO: replace magic number here, and in tests.
		time.Sleep(250 * time.Millisecond)

	}
}

func (wf *WatchedFile) scan() {
	currentCheck := wf.currentChecksum()

	if currentCheck == 0 {
		wf.SetBeingWatched(false)
		log.Debug("File Stat error")
		return // try again later?
	}

	wf.SetBeingWatched(true)

	if wf.checksum != currentCheck {
		go wf.OnFileChange(wf)
		wf.checksum = currentCheck
	}
}

func (wf *WatchedFile) currentChecksum() uint32 {

	file, err := ioutil.ReadFile(wf.Path)
	if err != nil {
		return 0
	}

	return crc32.ChecksumIEEE(file)
}

func (wf *WatchedFile) StartListener() {
	wf.keepPolling = true
	go wf.listen()
}
