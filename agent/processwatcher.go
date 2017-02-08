package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
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
type systemLibrary struct {
	Path           string
	PackageName    string
	PackageVersion string
	Modified       time.Time
}

// "map" in the colloquial sense
type systemState struct {
	processes systemProcesses
	libraries processLibraryMap
}

type processLibraryMap map[string]systemLibrary

type watchedProcess struct {
	ProcessStartedAt time.Time
	ProcessLibraries []processLibrary
	Outdated         bool
	Pid              int
	Command          string
}

type processLibrary struct {
	libraryPath string // point to a systemLibrary
	Outdated    bool
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
			libraryToString(buffer, lib.Outdated, ss.libraries[lib.libraryPath])
		}
	}
	return buffer.String()
}

func libraryToString(buffer *bytes.Buffer, outdated bool, lib systemLibrary) {
	if outdated {
		buffer.WriteString("Outdated: yes")
	} else {
		buffer.WriteString("Outdated:  no")
	}

	buffer.WriteString(fmt.Sprintf(", Path: %v, ", lib.Path))
	buffer.WriteString(fmt.Sprintf("Package: %s-%s, ", lib.PackageName, lib.PackageVersion))
	buffer.WriteString(fmt.Sprintf("Modified: %v", lib.Modified))

	return
}

func (ss *systemState) MarshalJSON() ([]byte, error) {
	libraries := make(map[string]interface{}, len(ss.libraries))
	for path, lib := range ss.libraries {
		libraries[path] = libToMap(lib)
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
			"outdated":     lib.Outdated,
			"library_path": lib.libraryPath,
		}
	}

	return procLibs
}

func libToMap(lib systemLibrary) map[string]interface{} {
	return map[string]interface{}{
		"path":            lib.Path,
		"modified":        lib.Modified,
		"package_name":    lib.PackageName,
		"package_version": lib.PackageVersion,
	}
}

func remove(s []libspector.Library, i int) []libspector.Library {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (pw *processWatcher) acquireState() *systemState {
	lsProcs, err := pw.processes()
	if err != nil {
		panic(fmt.Errorf("Couldn't load processes: %s", err))
	}

	ss := systemState{
		processes: make(systemProcesses, 0, len(lsProcs)),
		libraries: make(processLibraryMap, 0), // ¯\_(ツ)_/¯
	}

	rejects := map[string]bool{}

	for _, lsProc := range lsProcs {
		started, err := lsProc.Started()
		if err != nil {
			log.Debugf("PID %d is not running, skipping", lsProc.PID())
			continue
		}

		command, err := lsProc.Command()
		if err != nil {
			log.Debugf("Can't read command line for PID %d: %v", lsProc.PID(), err)
			// fall through, we can live without this (?)
		}

		spectorLibs, err := lsProc.Libraries()
		if err != nil {
			if os.Getuid() != 0 && os.Geteuid() != 0 {
				log.Debugf("Cannot examine libs for PID %d, with UID:%d, EUID:%d",
					lsProc.PID(), os.Getuid(), os.Geteuid())
				continue
			}

			if strings.Contains(err.Error(), "42") {
				// process went away
				log.Debugf("Cannot examine libs for PID %d, process disappeared",
					lsProc.PID())
				continue
			}

			// otherwise barf
			panic(fmt.Errorf("Couldn't load libs for process %v: %s", lsProc, err))
		}

		wp := watchedProcess{
			ProcessStartedAt: started,
			Pid:              lsProc.PID(),
			ProcessLibraries: make([]processLibrary, 0, len(spectorLibs)),
			Outdated:         false,
			Command:          command,
		}

		for _, spectorLib := range spectorLibs {
			path := spectorLib.Path()
			if rejects[path] {
				log.Debugf("Already rejected %v", path)
				continue
			}

			if _, ok := ss.libraries[path]; !ok {
				sysLib, err := NewSystemLibrary(spectorLib)
				if err != nil {
					log.Debugf("error introspecting system lib %s, %v; removing...", path, err)
					rejects[path] = true
					continue
				}

				ss.libraries[path] = sysLib
			}

			lib := processLibrary{
				libraryPath: path,
				Outdated:    spectorLib.Outdated(lsProc),
			}

			if lib.Outdated && !wp.Outdated {
				wp.Outdated = true
			}

			wp.ProcessLibraries = append(wp.ProcessLibraries, lib)
		}

		ss.processes = append(ss.processes, wp)
	}

	return &ss
}

func NewSystemLibrary(lib libspector.Library) (sysLib systemLibrary, err error) {
	path := lib.Path()

	modified, err := lib.Modified()
	if err != nil {
		return
	}

	pkg, err := lib.Package()
	if err != nil {
		return
	}

	pkgName := pkg.Name()
	pkgVersion := pkg.Version()

	sysLib = systemLibrary{
		Path:           path,
		Modified:       modified,
		PackageName:    pkgName,
		PackageVersion: pkgVersion,
	}

	return
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
			"system_state": state,
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
