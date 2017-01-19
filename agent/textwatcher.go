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

type TextWatcher interface {
	Start()
	Stop()
	Contents() ([]byte, error)
	Path() string
	Kind() string
	MarshalJSON() ([]byte, error)
}

type textWatcher struct {
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

// File watchers track changes in the contents of a file
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

	watcher := &textWatcher{
		path:      path,
		OnChange:  callback,
		kind:      kind,
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
	}
	watcher.contents = watcher.FileContents

	// Do a scan off the bat so we get a checksum, and PUT the file
	watcher.scan()
	return watcher
}

// Process watchers track changes in the output of a command
func NewCommandOutputWatcher(process string, callback ChangeHandler) Watcher {
	splat := strings.Split(process, " ")
	name := splat[0]
	args := splat[1:]

	watcher := &textWatcher{
		path:      process,
		OnChange:  callback,
		kind:      "centos",
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
		CmdName:   name,
		CmdArgs:   args,
	}
	watcher.contents = watcher.ProcessContents

	watcher.scan()
	return watcher
}

func (tw *textWatcher) MarshalJSON() ([]byte, error) {
	tw.Lock()
	defer tw.Unlock()
	ret, err := json.Marshal(map[string]interface{}{
		"path":          tw.Path(),
		"kind":          tw.Kind(),
		"updated-at":    tw.UpdatedAt,
		"being-watched": tw.BeingWatched,
		"crc":           tw.Checksum})
	return ret, err
}

func (wt *textWatcher) Kind() string {
	return wt.kind
}

func (wt *textWatcher) Path() string {
	return wt.path
}

func (wt *textWatcher) KeepPolling() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.keepPolling
}

func (wt *textWatcher) Start() {
	// log.Debug("Listening to: %s", wt.Path())
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = true
	go wt.listen()
}

func (wt *textWatcher) Stop() {
	// log.Debug("No longer listening to: %s", wt.Path())
	wt.Lock()
	defer wt.Unlock()
	wt.keepPolling = false
}

func (wt *textWatcher) GetBeingWatched() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.BeingWatched
}

func (wt *textWatcher) SetBeingWatched(bw bool) {
	wt.Lock()
	wt.BeingWatched = bw
	wt.Unlock()
}

// since on init the checksum never match, we always trigger an OnChange when we
// boot up
func (wt *textWatcher) scan() {
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

func (wt *textWatcher) currentChecksum() uint32 {
	file, err := wt.Contents()
	if err != nil {
		return 0
	}

	return crc32.ChecksumIEEE(file)
}

func (wt *textWatcher) listen() {
	for wt.KeepPolling() {
		wt.scan()
		time.Sleep(wt.pollSleep)
	}
}

func (wt *textWatcher) Contents() ([]byte, error) {
	return wt.contents()
}

func (wt *textWatcher) FileContents() ([]byte, error) {
	// log.Debug("####### file contents for %s!", wt.Path())
	return ioutil.ReadFile(wt.Path())
}

func (wt *textWatcher) ProcessContents() ([]byte, error) {
	// log.Debug("####### process contents!")
	cmd := exec.Command(wt.CmdName, wt.CmdArgs...)
	out, err := cmd.Output()

	return out, err
}
