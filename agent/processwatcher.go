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
	State() *processMap
}

// Types

// Singleton control object
type processWatcher struct {
	sync.Mutex
	keepPolling  bool
	UpdatedAt    time.Time
	OnChange     ChangeHandler
	pollSleep    time.Duration
	BeingWatched bool
	match        string
	state        *processMap
}

// process objects with references to systemLibraries
type systemProcesses []watchedProcess

// indexed list of libraries
type systemLibraries []libspector.Library

// This is how the data structure should look

// var someLib1 libspector.Library
// var someLib2 libspector.Library
// var processStartTime time.Time
// var updatedAt time.Time
// var callback = func(w Watcher) { return }

// var watcher = processWatcher{
// 	keepPolling: true,
// 	UpdatedAt:   updatedAt,
// 	OnChange:    callback,
// 	pollSleep:   env.PollSleep,
// 	match:       "",
// 	state: &processMap{
// 		processes: &systemProcesses{
// 			watchedProcess{
// 				ProcessStarted: processStartTime,
// 				Libraries: []library{
// 					processLibrary{
// 						Outdated:     true,
// 						libraryIndex: 2, // somelib2
// 					},
// 				},
// 				Outdated: true,
// 				Pid:      5,
// 				Command:  "test_script.sh 1 2 3",
// 			},
// 			// ...
// 		},
// 		libraries: &systemLibraries{
// 			someLib1,
// 			someLib2,
// 			// ...
// 		},
// 	},
// }

// "map" in the colloquial sense
type processMap struct {
	processes systemProcesses
	libraries systemLibraries
}

type watchedProcess struct {
	ProcessStarted time.Time
	Libraries      []processLibrary
	Outdated       bool
	Pid            int
	Command        string
}

type processLibrary struct {
	libraryIndex int // point to a systemLibrary
	Outdated     bool
}

// Methods

// processMap

func (pm *processMap) String() string {
	buffer := bytes.NewBufferString("Watched Processes:\n\n")

	for _, proc := range pm.processes {
		buffer.WriteString(fmt.Sprintf("PID: %d", proc.Pid))

		if proc.Outdated {
			buffer.WriteString(", is running outdated lib(s)")
		}

		buffer.WriteString(fmt.Sprintf("\nCommand: %s", proc.Command))

		for _, lib := range proc.Libraries {
			buffer.WriteString("\n")
			libraryToString(buffer, lib.Outdated, pm.libraries[lib.libraryIndex])
		}
	}
	return buffer.String()
}

func libraryToString(buffer *bytes.Buffer, outdated bool, lib libspector.Library) {
	if outdated {
		buffer.WriteString("Outdated: yes")
	} else {
		buffer.WriteString("Outdated:  no")
	}
	buffer.WriteString(fmt.Sprintf(", Path: %v, ", lib.Path()))
	if pkg, err := lib.Package(); err != nil {
		buffer.WriteString(fmt.Sprintf("Package error: %v", err))
	} else {
		buffer.WriteString(fmt.Sprintf("Package: %s-%s, ", pkg.Name(), pkg.Version()))
	}
	if modified, err := lib.Modified(); err != nil {
		buffer.WriteString(fmt.Sprintf("Modified error: %v", err))
	} else {
		buffer.WriteString(fmt.Sprintf("Modified: %v", modified))
	}
	return
}

func (pm *processMap) MarshalJSON() ([]byte, error) {
	libraries := make([]map[string]interface{}, len(pm.libraries))
	for i, lib := range pm.libraries {
		libraries[i] = libToMap(lib)
	}

	processes := make([]map[string]interface{}, len(pm.processes))
	for i, proc := range pm.processes {
		processes[i] = map[string]interface{}{
			"started":   proc.ProcessStarted,
			"libraries": procLibsToMapArray(proc.Libraries),
			"outdated":  proc.Outdated,
			"pid":       proc.Pid,
			"name":      proc.Command,
		}
	}

	return json.Marshal(map[string]interface{}{
		"processes": processes,
		"libraries": libraries,
	})
}

func procLibsToMapArray(libs []processLibrary) []map[string]interface{} {
	procLibs := make([]map[string]interface{}, len(libs))

	for i, lib := range libs {
		procLibs[i] = map[string]interface{}{
			"outdated":      lib.Outdated,
			"library_index": lib.libraryIndex,
		}
	}

	return procLibs
}

func libToMap(lib libspector.Library) map[string]interface{} {
	path := lib.Path()

	modified, err := lib.Modified()
	if err != nil {
		log.Warningf("error retrieving modification date for lib %s, %v", path, err)
		return nil
	}

	pkgName := ""
	pkgVersion := ""
	if pkg, err := lib.Package(); err != nil {
		log.Warningf("error retrieving package name for lib %s, %v", path, err)
	} else {
		pkgName = pkg.Name()
		pkgVersion = pkg.Version()
	}

	return map[string]interface{}{
		"path":            path,
		"modified":        modified,
		"package_name":    pkgName,
		"package_version": pkgVersion,
	}
}

func (pm *processMap) findLibraryIndex(path string) int {
	for i, library := range pm.libraries {
		if library.Path() == path {
			return i
		}
	}

	return -1
}

func (pm *processMap) maybeAddLibrary(lib libspector.Library) int {
	path := lib.Path()
	index := pm.findLibraryIndex(path)

	if index < 0 {
		pm.libraries = append(pm.libraries, lib)
		index = pm.findLibraryIndex(path)
	}

	return index
}

func (wt *processWatcher) acquireState() *processMap {
	procs, err := wt.processes()
	if err != nil {
		panic(fmt.Errorf("Couldn't load processes: %s", err))
	}

	pm := processMap{
		processes: make(systemProcesses, len(procs)),
		libraries: make(systemLibraries, 0), // ¯\_(ツ)_/¯
	}

	for _, proc := range procs {
		started, err := proc.Started()
		if err != nil {
			log.Infof("PID %d is not running, skipping", proc.PID())
			continue
		}

		command, err := proc.Command()
		if err != nil {
			log.Infof("Can't read command line for PID %d: %v", proc.PID(), err)
			// fall through, we can live without this (?)
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
			Libraries:      make([]processLibrary, len(spectorLibs)),
			Outdated:       false,
			Command:        command,
		}

		for i, spectorLib := range spectorLibs {
			libraryIndex := pm.maybeAddLibrary(spectorLib)
			lib := processLibrary{
				libraryIndex: libraryIndex,
				Outdated:     spectorLib.Outdated(proc),
			}
			if lib.Outdated && !wp.Outdated {
				wp.Outdated = true
			}
			wp.Libraries[i] = lib
		}

		pm.processes = append(pm.processes, wp)
	}

	return &pm
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

func NewAllProcessWatcher(callback ChangeHandler) Watcher {
	return NewProcessWatcher("", callback)
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

func (wt *processWatcher) State() *processMap {
	wt.Lock()
	defer wt.Unlock()
	if wt.state == nil {
		wt.state = wt.acquireState()
	}
	return wt.state
}

func (wt *processWatcher) processes() (procs []libspector.Process, err error) {
	if wt.match == "" {
		procs, err = libspector.AllProcesses()
	} else {
		procs, err = libspector.FindProcess(wt.match)
	}
	return
}

func (wt *processWatcher) scan() {
	wt.Lock()
	wt.state = wt.acquireState()
	wt.Unlock()
	// "OnChange" seems like a misnomer here.
	go wt.OnChange(wt)
}

func (wt *processWatcher) listen() {
	for wt.KeepPolling() {
		wt.scan()
		time.Sleep(wt.pollSleep)
	}
}

func singleServingWatcher() *processWatcher {
	return &processWatcher{
		match:     "",
		OnChange:  func(w Watcher) { return },
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
	}
}

func DumpProcessMap() {
	watcher := singleServingWatcher()
	fmt.Printf("%s\n", watcher.State().String())
}

func DumpJsonProcessMap() {
	watcher := singleServingWatcher()
	json, err := watcher.State().MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(json))
}
