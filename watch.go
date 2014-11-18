package agent

import (
	"fmt"

	"github.com/mveytsman/canary-agent/data"
	"github.com/mveytsman/canary-agent/parsers/gemfile"

	fsnotify "gopkg.in/fsnotify.v1"
)

func NewWatcher() *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		lg.Fatal(err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				lg.Info("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					lg.Info("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				lg.Info("error:", err)

			}
		}
	}()
	return watcher
}

func (a *Agent) AddGemfile(appName string, path string) {
	gemfileModel := &data.Gemfile{}

	a.db.FirstOrCreate(gemfileModel, data.Gemfile{AppName: appName, Path: path})

	gf, err := gemfile.ParseGemfile(path)
	if err != nil {
		//TODO handle error more gracefully
		lg.Fatal(err)
	}

	fmt.Println(gf)
	for _, spec := range gf.Specs {
		gem := &data.Gem{}
		a.db.Where(data.Gem{Name: spec.Name, GemfileId: gemfileModel.Id}).FirstOrInit(gem)
		if a.db.NewRecord(gem) {
			gem.Version = spec.Version
			lg.Info("Installed a new gem " + gem.Name + " in Gemfile " + gemfileModel.AppName)
			a.db.Save(gem)
		} else if gem.Version != spec.Version {
			lg.Info("Changed gem " + gem.Name + " in Gemfile " + gemfileModel.AppName + " from " + gem.Version + "to " + spec.Version)
			gem.Version = spec.Version
			a.db.Save(gem)
		} else {
			lg.Info("Did not change " + gem.Name + " in Gemfile " + gemfileModel.AppName)
		}
	}
	//TODO: record gems that are deleted

	err = a.watcher.Add(path)
	if err != nil {
		lg.Fatal(err)
	}
}

// TODO: make this a finalizer? :(
func (a *Agent) Close() {
	lg.Info("closing watcher")
	a.watcher.Close()
}
