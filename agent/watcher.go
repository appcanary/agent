package agent

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hash/crc32"
)

type ChangeHandler func(Watcher)

type Watcher interface {
	Start()
	Stop()
	Contents() ([]byte, error)
	Path() string
	Kind() string
	MarshalJSON() ([]byte, error)
}

type watcher struct {
	sync.Mutex
	keepPolling  bool
	kind         string
	path         string
	UpdatedAt    time.Time
	BeingWatched bool
	OnChange     ChangeHandler
	Checksum     uint32
	CmdName      string
	CmdArgs      []string
	contents     func() ([]byte, error)
	pollSleep    time.Duration
}

type Watchers []Watcher

// TODO: time.Now() needs to be called whenever it updates
func NewFileWatcherWithHook(path string, callback ChangeHandler) Watcher {
	w := NewFileWatcher(path, callback)
	return w
}

func NewProcessWatcherWithHook(path string, callback ChangeHandler) Watcher {
	w := NewProcessWatcher(path, callback)
	return w
}

// only used for tests
func NewFileWatcher(path string, callback ChangeHandler) Watcher {
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
	file := &watcher{path: path, OnChange: callback, kind: kind, UpdatedAt: time.Now(), pollSleep: env.PollSleep}
	file.contents = file.FileContents

	// Do a scan off the bat so we get a checksum, and PUT the file
	file.scan()
	return file
}

func NewProcessWatcher(process string, callback ChangeHandler) Watcher {

	splat := strings.Split(process, " ")
	name := splat[0]
	args := splat[1:]

	watcher := &watcher{path: process, OnChange: callback, kind: "centos", UpdatedAt: time.Now(), pollSleep: env.PollSleep, CmdName: name, CmdArgs: args}
	watcher.contents = watcher.ProcessContents

	watcher.scan()
	return watcher
}

func (wf *watcher) MarshalJSON() ([]byte, error) {
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

func (wt *watcher) Kind() string {
	return wt.kind
}

func (wt *watcher) Path() string {
	return wt.path
}

func (wt *watcher) KeepPolling() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.keepPolling
}

func (wt *watcher) Start() {
	// log.Debug("Listening to: %s", wt.Path())
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = true
	go wt.listen()
}

func (wt *watcher) Stop() {
	// log.Debug("No longer listening to: %s", wt.Path())
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = false
}

func (wt *watcher) GetBeingWatched() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.BeingWatched
}

func (wt *watcher) SetBeingWatched(bw bool) {
	wt.Lock()
	wt.BeingWatched = bw
	wt.Unlock()
}

// since on init the checksum never match,
// we always trigger an OnChange when we boot up
func (wt *watcher) scan() {
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
		go wt.OnChange(wt)
		wt.Checksum = currentCheck
	}
}

func (wt *watcher) currentChecksum() uint32 {

	file, err := wt.Contents()
	if err != nil {
		return 0
	}

	return crc32.ChecksumIEEE(file)
}

func (wt *watcher) listen() {
	for wt.KeepPolling() {
		wt.scan()
		time.Sleep(wt.pollSleep)
	}
}

func (wt *watcher) Contents() ([]byte, error) {
	return wt.contents()
}

func (wt *watcher) FileContents() ([]byte, error) {
	// log.Debug("####### file contents for %s!", wt.Path())
	return ioutil.ReadFile(wt.Path())
}

func (wt *watcher) ProcessContents() ([]byte, error) {
	// log.Debug("####### process contents!")
	cmd := exec.Command(wt.CmdName, wt.CmdArgs...)
	out, err := cmd.Output()

	return out, err
}
