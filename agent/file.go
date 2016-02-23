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

type FileChangeHandler func(Watcher)

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
	kind         string            `json:"-"`
	path         string            `json:"-"`
	UpdatedAt    time.Time         `json:"updated-at"`
	BeingWatched bool              `json:"being-watched"`
	OnFileChange FileChangeHandler `json:"-"`
	Checksum     uint32            `json:"crc"`
	CmdName      string
	CmdArgs      []string
	contents     func() ([]byte, error)
}

type Watchers []Watcher

// TODO: time.Now() needs to be called whenever it updates
func NewFileWatcherWithHook(path string, callback FileChangeHandler) Watcher {
	w := NewFileWatcher(path, callback)
	w.Start()
	return w
}

func NewProcessWatcherWithHook(path string, callback FileChangeHandler) Watcher {
	w := NewProcessWatcher(path, callback)
	w.Start()
	return w
}

// only used for tests
func NewFileWatcher(path string, callback FileChangeHandler) Watcher {
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
	file := &watcher{path: path, OnFileChange: callback, kind: kind, UpdatedAt: time.Now()}
	file.contents = file.FileContents

	// Do a scan off the bat so we get a checksum, and PUT the file
	file.scan()
	return file
}

func NewProcessWatcher(process string, callback FileChangeHandler) Watcher {

	splat := strings.Split(process, " ")
	name := splat[0]
	args := splat[1:]

	watcher := &watcher{path: process, OnFileChange: callback, kind: "centos", UpdatedAt: time.Now(), CmdName: name, CmdArgs: args}
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
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = true
	go wt.listen()
}

// TODO: solve data race issue
func (wt *watcher) Stop() {
	log.Debug("No longer listening to: %s", wt.Path())
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
		go wt.OnFileChange(wt)
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
		time.Sleep(env.PollSleep)

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
