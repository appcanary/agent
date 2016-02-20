package agent

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hash/crc32"
)

const (
	WatchedProcess = 1
	WatchedFile    = 2
)

type FileChangeHandler func(Watcher)

type Watcher interface {
	Start()
	Stop()
	Contents() ([]byte, error)
	Path() string
	Kind() string
	MarshalJSON() ([]byte, error)
}

type WatchedThing struct {
	sync.Mutex
	keepPolling  bool
	kind         string            `json:"-"`
	path         string            `json:"-"`
	UpdatedAt    time.Time         `json:"updated-at"`
	BeingWatched bool              `json:"being-watched"`
	OnFileChange FileChangeHandler `json:"-"`
	Checksum     uint32            `json:"crc"`
	Name         string
	Args         []string
	Klass        int
}

type Watchers []Watcher

// TODO: time.Now() needs to be called whenever it updates
func NewWatcherWithHook(path string, callback FileChangeHandler, klass int) Watcher {
	var w Watcher
	switch klass {
	case WatchedFile:
		w = NewWatchedFile(path, callback)
	case WatchedProcess:
		w = NewWatchedProcess(path, callback)
	}
	w.Start()
	return w
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
	file := &WatchedThing{path: path, OnFileChange: callback, kind: kind, UpdatedAt: time.Now(), Klass: WatchedFile}

	// Do a scan off the bat so we get a checksum, and PUT the file
	file.scan()
	return file
}

func NewWatchedProcess(process string, callback FileChangeHandler) Watcher {

	splat := strings.Split(process, " ")
	name := splat[0]
	args := splat[1:]

	watcher := &WatchedThing{path: process, OnFileChange: callback, kind: "process", UpdatedAt: time.Now(), Name: name, Args: args, Klass: WatchedProcess}

	watcher.scan()
	return watcher
}

func (wf *WatchedThing) MarshalJSON() ([]byte, error) {
	wf.Lock()
	defer wf.Unlock()
	ret, err := json.Marshal(map[string]interface{}{
		"path":          wf.Path(),
		"kind":          wf.Kind(),
		"updated-at":    wf.UpdatedAt,
		"being-watched": wf.BeingWatched,
		"crc":           wf.Checksum})
	return ret, err
}

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
	log.Debug("No longer listening to: %s", wt.Path())
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
	switch wt.Klass {
	case WatchedFile:
		return wt.FileContents()
	case WatchedProcess:
		return wt.ProcessContents()
	}
	return nil, errors.New("Invalid WatchedThing class")
}

func (wt *WatchedThing) scan() {
	// log.Debug("wt: Check for %s", wt.Path())
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

func (wt *WatchedThing) FileContents() ([]byte, error) {
	// log.Debug("####### file contents for %s!", wt.Path())
	return ioutil.ReadFile(wt.Path())
}

func (wt *WatchedThing) listen() {
	for wt.KeepPolling() {

		wt.scan()
		time.Sleep(POLL_SLEEP)

	}
}

func (wt *WatchedThing) ProcessContents() ([]byte, error) {
	// log.Debug("####### process contents!")
	cmd := exec.Command(wt.Name, wt.Args...)
	out, err := cmd.Output()

	return out, err
}
