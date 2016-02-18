package agent

import (
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"hash/crc32"
)

type FileChangeHandler func(Watcher)

type Watcher interface {
	Start()
	Stop()
	Contents() ([]byte, error)
	Path() string
	Kind() string
}

type WatchedThing struct {
	sync.Mutex
	keepPolling  bool
	kind         string            `json:"kind"`
	path         string            `json:"path"`
	UpdatedAt    time.Time         `json:"updated-at"`
	BeingWatched bool              `json:"being-watched"`
	OnFileChange FileChangeHandler `json:"-"`
	Checksum     uint32            `json:"crc"`
}

type WatchedFile struct {
	*WatchedThing
}

type WatchedFiles []*WatchedFile
type WatchedThings []*WatchedThing
type Watchers []Watcher

// TODO: time.Now() needs to be called whenever it updates
func NewWatchedFileWithHook(path string, callback FileChangeHandler) Watcher {
	file := NewWatchedFile(path, callback)
	file.Start()
	return file
}

// only used for tests
func NewWatchedFile(path string, callback FileChangeHandler) Watcher {
	var kind string
	filename := filepath.Base(path)
	switch filename {
	case "Gemfile.lock":
		kind = "gemfile"
	case "available":
		//todo support debian
		kind = "ubuntu"
	case "status":
		kind = "ubuntu"
	}
	file := &WatchedFile{&WatchedThing{path: path, OnFileChange: callback, kind: kind, UpdatedAt: time.Now()}}

	// Do a scan off the bat so we get a checksum, and PUT the file
	file.scan()
	return file
}

// func (wf *WatchedFile) MarshalJson() ([]byte, error) {
// 	wf.Lock()
// 	defer wf.Unlock()
// 	ret, err := json.Marshal(interface{}(wf))
// 	return ret, err
// }

func (wt *WatchedThing) Kind() string {
	return wt.kind
}

func (wt *WatchedThing) Path() string {
	return wt.path
}

func (wt *WatchedThing) KeepPolling() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.keepPolling
}

func (wt *WatchedThing) Start() {
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = true
	go wt.listen()
}

// TODO: solve data race issue
func (wt *WatchedThing) Stop() {
	log.Debug("No longer listening to: %s", wt.Path)
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = false
}

func (wt *WatchedThing) GetBeingWatched() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.BeingWatched
}

func (wt *WatchedThing) SetBeingWatched(bw bool) {
	wt.Lock()
	wt.BeingWatched = bw
	wt.Unlock()
}

func (wt *WatchedThing) Contents() ([]byte, error) {
	return ioutil.ReadFile(wt.Path())
}

func (wt *WatchedThing) listen() {
	for wt.KeepPolling() {

		wt.scan()
		time.Sleep(POLL_SLEEP)

	}
}

func (wt *WatchedThing) scan() {
	// log.Debug("wt: Check.")
	currentCheck := wt.currentChecksum()

	if currentCheck == 0 {
		// log.Debug("wt: checksum fail.")
		// there was some error reading the file.
		// try again later?
		wt.SetBeingWatched(false)
		return
	}

	wt.SetBeingWatched(true)

	if wt.Checksum != currentCheck {
		go wt.OnFileChange(wt)
		wt.Checksum = currentCheck
	}
}

func (wt *WatchedThing) currentChecksum() uint32 {

	file, err := wt.Contents()
	if err != nil {
		return 0
	}

	return crc32.ChecksumIEEE(file)
}
