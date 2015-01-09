package agent

import (
	"gopkg.in/fsnotify.v1"
)

// Helpers for fsnotify
// TODO move this into a package
// helpers for fsnotify
func isCreate(o fsnotify.Op) bool {
	return (o&fsnotify.Create == fsnotify.Create)
}

func isRename(o fsnotify.Op) bool {
	return (o&fsnotify.Rename == fsnotify.Rename)
}
