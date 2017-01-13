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
	processStarted time.Time
	libraries      []libspector.Library
	outdated       bool
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

func (wt *processWatcher) scan() {
	procs, err := libspector.FindProcess(wt.match) // or AllProcesses?
	if err != nil {
		panic(err)
	}

	var state watchedState
	for _, proc := range procs {
		started, err := proc.Started()
		if err != nil {
			log.Infof("PID %d is not running, skipping", proc.PID())
			continue
		}

		if libs, err := proc.Libraries(); err != nil {
			panic(fmt.Errorf("Couldn't load libs for process %v: %s", proc, err))
		} else {
			process := watchedProcess{
				libraries:      libs,
				processStarted: started,
			}

			for _, lib := range libs {
				if lib.Outdated(proc) {
					process.outdated = true
					break
				}
			}

			state[proc.PID()] = process
		}
	}

	// TODO so the big question is, the whole list, or just the partial list,
	// and do we ship all of it. Anyway, "OnChange" seems like a misnomer here.
	go wt.OnChange(wt)
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
