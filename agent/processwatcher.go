package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/appcanary/libspector"
)

type ProcessWatcher interface {
	Start()
	Stop()
	Match() string
	StateJson() []byte
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
	stateJson    []byte
	checksum     uint32
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
// 	match:       "*",
//  // this is no longer true, instead we just keep the marshaled json, but this
//  // a helpful structure doc
// 	state: &systemState{
// 		processes: &systemProcesses{
// 			watchedProcess{
// 				ProcessStartedAt: processStartTime,
// 				ProcessLibraries: []processLibrary{
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
type systemState struct {
	processes         systemProcesses
	libraries         systemLibraries
	processLibraryMap processLibraryMap
}

type processLibraryMap map[string]int

type watchedProcess struct {
	ProcessStartedAt time.Time
	ProcessLibraries []processLibrary
	Outdated         bool
	Pid              int
	Command          string
}

type processLibrary struct {
	libraryIndex int // point to a systemLibrary
	Outdated     bool
}

func (ss *systemState) String() string {
	buffer := bytes.NewBufferString("Watched Processes:\n\n")

	for _, proc := range ss.processes {
		buffer.WriteString(fmt.Sprintf("PID: %d", proc.Pid))

		if proc.Outdated {
			buffer.WriteString(", is running outdated lib(s)")
		}

		buffer.WriteString(fmt.Sprintf("\nCommand: %s", proc.Command))

		for _, lib := range proc.ProcessLibraries {
			buffer.WriteString("\n")
			libraryToString(buffer, lib.Outdated, ss.libraries[lib.libraryIndex])
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
		log.Warningf("Error reading package info for %s: %v", lib.Path(), err)
		buffer.WriteString("Package: unknown, ")
	} else {
		buffer.WriteString(fmt.Sprintf("Package: %s-%s, ", pkg.Name(), pkg.Version()))
	}

	if modified, err := lib.Modified(); err != nil {
		log.Warningf("Error reading modification date: %v", err)
		buffer.WriteString("Modified: unknown")
	} else {
		buffer.WriteString(fmt.Sprintf("Modified: %v", modified))
	}

	return
}

func (ss *systemState) MarshalJSON() ([]byte, error) {
	libraries := make([]map[string]interface{}, len(ss.libraries))
	for i, lib := range ss.libraries {
		libraries[i] = libToMap(lib)
	}

	processes := make([]map[string]interface{}, len(ss.processes))
	for i, proc := range ss.processes {
		processes[i] = map[string]interface{}{
			"started":   proc.ProcessStartedAt,
			"libraries": procLibsToMapArray(proc.ProcessLibraries),
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

func (sl processLibraryMap) findLibraryIndex(path string) int {
	if index, ok := sl[path]; ok {
		return index
	}

	return -1
}

func (ss *systemState) findLibraryIndex(path string) int {
	return ss.processLibraryMap.findLibraryIndex(path)
}

func (ss *systemState) addLibrary(lib libspector.Library) int {
	path := lib.Path()
	index := ss.findLibraryIndex(path)

	if index < 0 {
		ss.libraries = append(ss.libraries, lib)
		index = len(ss.libraries) - 1
		ss.processLibraryMap[path] = index
	}

	return index
}

func (pw *processWatcher) acquireState() *systemState {
	lsProcs, err := pw.processes()
	if err != nil {
		panic(fmt.Errorf("Couldn't load processes: %s", err))
	}

	ss := systemState{
		processes:         make(systemProcesses, 0, len(lsProcs)),
		libraries:         make(systemLibraries, 0), // ¯\_(ツ)_/¯
		processLibraryMap: make(map[string]int, 0),
	}

	for _, lsProc := range lsProcs {
		started, err := lsProc.Started()
		if err != nil {
			log.Infof("PID %d is not running, skipping", lsProc.PID())
			continue
		}

		command, err := lsProc.Command()
		if err != nil {
			log.Infof("Can't read command line for PID %d: %v", lsProc.PID(), err)
			// fall through, we can live without this (?)
		}

		spectorLibs, err := lsProc.Libraries()
		if err != nil {
			if os.Getuid() != 0 && os.Geteuid() != 0 {
				log.Infof("Cannot examine libs for PID %d, with UID:%d, EUID:%d",
					lsProc.PID(), os.Getuid(), os.Geteuid())
				continue
			}

			// otherwise barf
			panic(fmt.Errorf("Couldn't load libs for process %v: %s", lsProc, err))
		}

		wp := watchedProcess{
			ProcessStartedAt: started,
			Pid:              lsProc.PID(),
			ProcessLibraries: make([]processLibrary, len(spectorLibs)),
			Outdated:         false,
			Command:          command,
		}

		for i, spectorLib := range spectorLibs {
			libraryIndex := ss.addLibrary(spectorLib)
			lib := processLibrary{
				libraryIndex: libraryIndex,
				Outdated:     spectorLib.Outdated(lsProc),
			}
			if lib.Outdated && !wp.Outdated {
				wp.Outdated = true
			}
			wp.ProcessLibraries[i] = lib
		}

		ss.processes = append(ss.processes, wp)
	}

	// to make the crc check more meaningful
	ss.sortSystemState()

	return &ss
}

func (ss *systemState) sortSystemState() {
	// make a copy in the original order
	oldSLs := make(systemLibraries, len(ss.libraries))
	copy(oldSLs, ss.libraries)

	// sort everything
	sort.Sort(ss.processes)
	sort.Sort(ss.libraries)

	for _, proc := range ss.processes {
		for _, procLib := range proc.ProcessLibraries {
			path := oldSLs[procLib.libraryIndex].Path()
			// find the new index
			procLib.libraryIndex = ss.findLibraryIndex(path)
		}
	}
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
	return NewProcessWatcher("*", callback)
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

func (pw *processWatcher) Start() {
	pw.Lock()
	pw.keepPolling = true
	pw.Unlock()
	go pw.listen()
}

func (pw *processWatcher) Stop() {
	pw.Lock()
	pw.keepPolling = false
	pw.Unlock()
}

func (pw *processWatcher) Match() string {
	return pw.match
}

func (pw *processWatcher) KeepPolling() bool {
	pw.Lock()
	defer pw.Unlock()
	return pw.keepPolling
}

func (pw *processWatcher) setStateAttribute() {
	state := pw.acquireState()

	json, err := json.Marshal(map[string]interface{}{
		"server": map[string]interface{}{
			"process_map": state,
		},
	})

	if err != nil {
		panic(err) // really shouldn't happen
	}

	pw.stateJson = json
}

func (pw *processWatcher) StateJson() []byte {
	pw.Lock()
	defer pw.Unlock()
	if pw.stateJson == nil {
		pw.setStateAttribute()
	}
	return pw.stateJson
}

func (pw *processWatcher) processes() (procs []libspector.Process, err error) {
	if pw.match == "*" {
		procs, err = libspector.AllProcesses()
	} else {
		procs, err = libspector.FindProcess(pw.match)
	}
	return
}

func (pw *processWatcher) scan() {
	pw.Lock()

	pw.setStateAttribute()

	newChecksum := crc32.ChecksumIEEE(pw.stateJson)
	changed := newChecksum != pw.checksum
	pw.checksum = newChecksum

	pw.Unlock() // ¯\_(ツ)_/¯

	if changed {
		go pw.OnChange(pw)
	}
}

func (pw *processWatcher) listen() {
	for pw.KeepPolling() {
		pw.scan()
		time.Sleep(pw.pollSleep)
	}
}

func singleServingWatcher() *processWatcher {
	return &processWatcher{
		match:     "*",
		OnChange:  func(w Watcher) { return },
		UpdatedAt: time.Now(),
		pollSleep: env.PollSleep,
	}
}

func DumpProcessMap() {
	watcher := singleServingWatcher()
	fmt.Printf("%s\n", watcher.acquireState().String())
}

func DumpJsonProcessMap() {
	watcher := singleServingWatcher()
	fmt.Printf("%s\n", string(watcher.StateJson()))
}

func (s systemProcesses) Len() int           { return len(s) }
func (s systemProcesses) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s systemProcesses) Less(i, j int) bool { return s[i].Pid < s[j].Pid }

func (s systemLibraries) Len() int           { return len(s) }
func (s systemLibraries) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s systemLibraries) Less(i, j int) bool { return s[i].Path() < s[j].Path() }
