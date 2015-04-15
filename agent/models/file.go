package models

import "gopkg.in/fsnotify.v1"

type File interface {
	GetPath() string
	Parse() interface{}
}

type WatchedFile struct {
	File
	Watcher *fsnotify.Watcher
}

type WatchedFiles []*WatchedFile

// TODO: make this a finalizer? :(
func (wf *WatchedFile) Close() {
	log.Info("closing watcher")
	wf.Watcher.Close()
}
