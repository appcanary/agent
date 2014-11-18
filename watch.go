package agent

import (
	"github.com/mveytsman/canary-agent/data"
	"github.com/mveytsman/canary-agent/parsers/gemfile"

	fsnotify "gopkg.in/fsnotify.v1"
)

type WatchedFile struct {
	agent   *Agent
	watcher *fsnotify.Watcher
	path    string
	appName string
}

func (a *Agent) NewWatchedFile(appName string, path string) *WatchedFile {
	wf := &WatchedFile{agent: a, appName: appName, path: path}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		lg.Fatal(err)
	}

	wf.watcher = watcher
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					//File is overwritten, we need to add a new watch to it
					//TODO: we need to be smart about pausing here
					err = wf.watcher.Add(path)
					if err != nil {
						lg.Fatal(err)
					}
				}
				// reread the gemfile
				wf.ReadGemfile()
			case err := <-watcher.Errors:
				lg.Info("error:", err)
			}
		}
	}()
	wf.ReadGemfile()
	err = wf.watcher.Add(path)
	if err != nil {
		lg.Fatal(err)
	}

	return wf
}

func (wf *WatchedFile) ReadGemfile() {
	db := wf.agent.db
	gemfileModel := &data.Gemfile{}

	db.FirstOrCreate(gemfileModel, data.Gemfile{AppName: wf.appName, Path: wf.path})

	gf, err := gemfile.ParseGemfile(wf.path)
	if err != nil {
		//TODO handle error more gracefully
		lg.Fatal(err)
	}

	for _, spec := range gf.Specs {
		gem := &data.Gem{}
		db.Where(data.Gem{Name: spec.Name, GemfileId: gemfileModel.Id}).FirstOrInit(gem)
		if db.NewRecord(gem) {
			gem.Version = spec.Version
			lg.Info("Installed a new gem " + gem.Name + " version " + gem.Version + " in Gemfile for app: " + gemfileModel.AppName)
			db.Save(gem)
		} else if gem.Version != spec.Version {
			lg.Info("Changed gem " + gem.Name + " in Gemfile for app: " + gemfileModel.AppName + " from " + gem.Version + "to " + spec.Version)
			gem.Version = spec.Version
			db.Save(gem)
		} else {
			//lg.Info("Did not change " + gem.Name + " in Gemfile " + gemfileModel.AppName)
		}
	}
	//TODO: record gems that are deleted
}

// TODO: make this a finalizer? :(
func (wf *WatchedFile) Close() {
	lg.Info("closing watcher")
	wf.watcher.Close()
}
