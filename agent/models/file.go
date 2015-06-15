package models

import (
	"io/ioutil"
	"sync"
	"time"

	"hash/crc32"

	"github.com/stateio/canary-agent/agent/umwelten"
)

const POLL_SLEEP = 250 * time.Millisecond

var log = umwelten.Log

type FileChangeHandler func(*WatchedFile)

type WatchedFile struct {
	sync.Mutex
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

// func (wf *WatchedFile) MarshalJson() ([]byte, error) {
// 	wf.Lock()
// 	defer wf.Unlock()
// 	ret, err := json.Marshal(interface{}(wf))
// 	return ret, err
// }

func (wf *WatchedFile) KeepPolling() bool {
	wf.Lock()
	defer wf.Unlock()
	return wf.keepPolling
}

func (wf *WatchedFile) StartListener() {
	wf.Lock()
	defer wf.Unlock()
	wf.keepPolling = true
	go wf.listen()
}

// TODO: solve data race issue
func (wf *WatchedFile) StopListening() {
	log.Debug("No longer listening to: %s", wf.Path)
	wf.Lock()
	defer wf.Unlock()
	wf.keepPolling = false
}

func (wf *WatchedFile) GetBeingWatched() bool {
	wf.Lock()
	defer wf.Unlock()
	return wf.BeingWatched
}

func (wf *WatchedFile) SetBeingWatched(bw bool) {
	wf.Lock()
	wf.BeingWatched = bw
	wf.Unlock()
}

func (wf *WatchedFile) Contents() ([]byte, error) {
	return ioutil.ReadFile(wf.Path)
}

func (wf *WatchedFile) listen() {
	for wf.KeepPolling() {

		wf.scan()
		// TODO: replace magic number here, and in tests.
		time.Sleep(POLL_SLEEP)

	}
}

func (wf *WatchedFile) scan() {
	currentCheck := wf.currentChecksum()

	if currentCheck == 0 {
		wf.SetBeingWatched(false)
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
