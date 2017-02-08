package agent

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/appcanary/libspector"
)

type ProcessWatcher interface {
	Start()
	Stop()
	Match() string
	State() *watchedState
}

type processWatcher struct {
	sync.Mutex
	keepPolling  bool
	UpdatedAt    time.Time
	OnChange     ChangeHandler
	pollSleep    time.Duration
	BeingWatched bool
	match        string
	state        *watchedState
}

type watchedState map[int]watchedProcess

type watchedProcess struct {
	ProcessStarted time.Time
	Libraries      []libspector.Library
	Outdated       bool
	Pid            int
}

func (wp *watchedProcess) MarshalJSON() ([]byte, error) {
	libs := make([]map[string]interface{}, len(wp.Libraries))

	for i, lib := range wp.Libraries {
		path := lib.Path()

		modified, err := lib.Modified()
		if err != nil {
			log.Warningf("error retrieving modification date for lib %s, %v", path, err)
			continue
		}

		pkg, err := lib.Package()
		if err != nil {
			log.Warningf("error retrieving package name for lib %s, %v", path, err)
		}

		libs[i] = map[string]interface{}{
			"path":     path,
			"modified": modified,
			"package":  pkg,
		}
	}

	return json.Marshal(map[string]interface{}{
		"started":   wp.ProcessStarted,
		"libraries": libs,
		"outdated":  wp.Outdated,
		"pid":       wp.Pid,
	})
}

func NewProcessWatcher(match string, callback ChangeHandler) Watcher {
	watcher := &processWatcher{
		match:     match,
		OnChange:  callback,
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
	}

	watcher.scan()
	return watcher
}

func NewAllProcessWatcher(callback ChangeHandler) Watcher {
	return NewProcessWatcher("", callback)
}

func (wt *processWatcher) Start() {
	wt.Lock()
	wt.keepPolling = true
	go wt.listen()
	wt.Unlock()
}

func (wt *processWatcher) Stop() {
	wt.Lock()
	wt.keepPolling = false
	wt.Unlock()
}

func (wt *processWatcher) Match() string {
	wt.Lock()
	defer wt.Unlock()
	return wt.match
}

func (wt *processWatcher) KeepPolling() bool {
	wt.Lock()
	defer wt.Unlock()
	return wt.keepPolling
}

func (wt *processWatcher) processes() (procs []libspector.Process, err error) {
	if wt.match == "" {
		procs, err = libspector.AllProcesses()
	} else {
		procs, err = libspector.FindProcess(wt.match)
	}
	return
}

func (wt *processWatcher) acquireState() *watchedState {
	procs, err := wt.processes()
	if err != nil {
		panic(fmt.Errorf("Couldn't load processes: %s", err))
	}

	state := make(watchedState)
	for _, proc := range procs {
		started, err := proc.Started()
		if err != nil {
			log.Infof("PID %d is not running, skipping", proc.PID())
			continue
		}

		libs, err := proc.Libraries()
		if err != nil {
			panic(fmt.Errorf("Couldn't load libs for process %v: %s", proc, err))
		}

		wp := watchedProcess{
			Libraries:      libs,
			ProcessStarted: started,
			Pid:            proc.PID(),
		}

		for _, lib := range libs {
			if lib.Outdated(proc) {
				wp.Outdated = true
				break
			}
		}

		state[proc.PID()] = wp
	}

	return &state
}

func (wt *processWatcher) scan() {
	wt.Lock()
	wt.state = wt.acquireState()
	// "OnChange" seems like a misnomer here.
	go wt.OnChange(wt)
	wt.Unlock()
}

func (wt *processWatcher) State() *watchedState {
	wt.Lock()
	defer wt.Unlock()
	return wt.state
}

func (wt *processWatcher) listen() {
	for wt.KeepPolling() {
		wt.scan()
		time.Sleep(wt.pollSleep)
	}
}
