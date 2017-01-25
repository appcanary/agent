package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

func (w watchedState) String() string {
	buffer := bytes.NewBufferString("Watched Processes:\n\n")
	for _, proc := range w {
		buffer.WriteString(fmt.Sprintf("%v\n", proc.String()))
	}
	return buffer.String()
}

type watchedProcess struct {
	ProcessStarted time.Time
	Libraries      []library
	Outdated       bool
	Pid            int
}

func (w *watchedProcess) String() string {
	buffer := bytes.NewBufferString(fmt.Sprintf("PID: %d", w.Pid))
	if w.Outdated {
		buffer.WriteString(", is running outdated lib(s)")
	}
	for _, lib := range w.Libraries {
		buffer.WriteString(fmt.Sprintf("\n%v", lib.String()))
	}
	return buffer.String()
}

type library struct {
	SpectorLib libspector.Library
	Outdated   bool
}

func (l *library) String() string {
	buffer := bytes.NewBufferString("")
	if l.Outdated {
		buffer.WriteString("Outdated: yes")
	} else {
		buffer.WriteString("Outdated:  no")
	}
	buffer.WriteString(fmt.Sprintf(", Path: %v, ", l.SpectorLib.Path()))
	if pkg, err := l.SpectorLib.Package(); err != nil {
		buffer.WriteString(fmt.Sprintf("Package error: %v", err))
	} else {
		buffer.WriteString(fmt.Sprintf("Package: %v, ", pkg))
	}
	if modified, err := l.SpectorLib.Modified(); err != nil {
		buffer.WriteString(fmt.Sprintf("Modified error: %v", err))
	} else {
		buffer.WriteString(fmt.Sprintf("Modified: %v", modified))
	}
	return buffer.String()
}

func (wp *watchedProcess) MarshalJSON() ([]byte, error) {
	libs := make([]map[string]interface{}, len(wp.Libraries))

	for i, lib := range wp.Libraries {
		path := lib.SpectorLib.Path()

		modified, err := lib.SpectorLib.Modified()
		if err != nil {
			log.Warningf("error retrieving modification date for lib %s, %v", path, err)
			continue
		}

		var fullPackage string
		if pkg, err := lib.SpectorLib.Package(); err != nil {
			log.Warningf("error retrieving package name for lib %s, %v", path, err)
		} else {
			fullPackage = fmt.Sprintf("%s-%s", pkg.Name(), pkg.Version())
		}

		libs[i] = map[string]interface{}{
			"path":     path,
			"modified": modified,
			"package":  fullPackage,
			"outdated": lib.Outdated, // in relation to this process
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

	// Don't scan from here, we just end up with two running at once
	return watcher
}

func (pw *processWatcher) MarshalJSON() ([]byte, error) {
	pw.Lock()
	defer pw.Unlock()
	return json.Marshal(map[string]interface{}{
		"match":         pw.Match(),
		"updated-at":    pw.UpdatedAt,
		"being-watched": pw.BeingWatched,
	})
}

func NewAllProcessWatcher(callback ChangeHandler) Watcher {
	return NewProcessWatcher("", callback)
}

func (wt *processWatcher) Start() {
	wt.Lock()
	wt.keepPolling = true
	wt.Unlock()
	go wt.listen()
}

func (wt *processWatcher) Stop() {
	wt.Lock()
	wt.keepPolling = false
	wt.Unlock()
}

func (wt *processWatcher) Match() string {
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

		spectorLibs, err := proc.Libraries()
		if err != nil {
			if os.Getuid() != 0 && os.Geteuid() != 0 {
				log.Infof("Cannot examine libs for PID %d, with UID:%d, EUID:%d",
					proc.PID(), os.Getuid(), os.Geteuid())
				continue
			}

			// otherwise barf
			panic(fmt.Errorf("Couldn't load libs for process %v: %s", proc, err))
		}

		wp := watchedProcess{
			ProcessStarted: started,
			Pid:            proc.PID(),
			Libraries:      make([]library, len(spectorLibs)),
			Outdated:       false,
		}

		for i, spectorLib := range spectorLibs {
			lib := library{
				SpectorLib: spectorLib,
				Outdated:   spectorLib.Outdated(proc),
			}
			if lib.Outdated && !wp.Outdated {
				wp.Outdated = true
			}
			wp.Libraries[i] = lib
		}

		state[proc.PID()] = wp
	}

	return &state
}

func (wt *processWatcher) scan() {
	wt.Lock()
	wt.state = wt.acquireState()
	wt.Unlock()
	// "OnChange" seems like a misnomer here.
	go wt.OnChange(wt)
}

func (wt *processWatcher) State() *watchedState {
	wt.Lock()
	defer wt.Unlock()
	if wt.state == nil {
		wt.state = wt.acquireState()
	}
	return wt.state
}

func (wt *processWatcher) listen() {
	for wt.KeepPolling() {
		wt.scan()
		time.Sleep(wt.pollSleep)
	}
}

func DumpProcessMap() {
	watcher := &processWatcher{
		match:     "",
		OnChange:  func(w Watcher) { return },
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
	}

	fmt.Fprintf(os.Stderr, "%s\n", watcher.State().String())
}
