package agent

type ChangeHandler func(Watcher)

type Watcher interface {
	Start()
	Stop()
}

type Watchers []Watcher
