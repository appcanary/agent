package agent

import (
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

func (wt *processWatcher) processes() []libspector.Process {
	var procs []libspector.Process
	var err error

	if wt.match == "" {
		procs, err = libspector.AllProcesses()
		fmt.Printf("all processes error: %v", err)
	} else {
		procs, err = libspector.FindProcess(wt.match)
		fmt.Printf("find processes error: %v", err)
	}

	if err != nil {
		panic(err)
	}

	return procs
}

func (wt *processWatcher) acquireState() *watchedState {
	procs := wt.processes()

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
