package agent

import (
	"github.com/mveytsman/canary-agent/parsers/gemfile"
	"gopkg.in/fsnotify.v1"
)

type File interface {
	GetPath() string
	Parse() interface{}
}

type WatchedFile struct {
	File
	Watcher *fsnotify.Watcher
}

type WatchedFiles []*WatchedFile

type Gemfile struct {
	Path string
}

func (g *Gemfile) GetPath() string {
	return g.Path
}

func (a *App) WatchFile(f File) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		lg.Fatal(err.Error())
	}

	wf := &WatchedFile{File: f, Watcher: watcher}
	go func() {
		//TODO 	defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					//File is overwritten, we need to add a new watch to it
					//TODO: we need to be smart about pausing here
					err = wf.Watcher.Add(wf.GetPath())
					if err != nil {
						lg.Fatal(err.Error())
					}
				}
				// reread the gemfile
				lg.Info("Rereading file: %s", wf.GetPath())
				wf.Parse()
				//a.Submit()
			case err := <-watcher.Errors:
				lg.Info("error:", err.Error())
			}
		}
	}()
	lg.Info("Reading file: %s", wf.GetPath())
	wf.Parse()
	//a.Submit()
	err = wf.Watcher.Add(wf.GetPath())
	if err != nil {
		lg.Fatal(err.Error())
	}

	//Add watched file to the apps list
	a.WatchedFiles = append(a.WatchedFiles, wf)
}

func (a *App) CloseWatches() {
	for _, wf := range a.WatchedFiles {
		wf.Close()
	}
}

func (g *Gemfile) Parse() interface{} {
	gf, err := gemfile.ParseGemfile(g.Path)
	if err != nil {
		//TODO handle error more gracefully
		lg.Fatal(err.Error())
	}
	return gf
}

// TODO: make this a finalizer? :(
func (wf *WatchedFile) Close() {
	lg.Info("closing watcher")
	wf.Watcher.Close()
}
