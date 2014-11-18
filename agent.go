package agent

import (
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
	"gopkg.in/fsnotify.v1"

	"github.com/mveytsman/canary-agent/data"
)

var lg = logging.MustGetLogger("app-canary")

type Agent struct {
	conf    *Conf
	watcher *fsnotify.Watcher
	db      gorm.DB
}

func NewAgent(confPath string) *Agent {
	agent := &Agent{}

	agent.conf = NewConfFromFile(confPath)
	agent.watcher = NewWatcher()
	agent.db = data.Initialize()

	// load the existing gemfiles yo'

	agent.AddGemfile(agent.conf.Ruby.Projects[0][0], agent.conf.Ruby.Projects[0][1]+"/Gemfile.lock")

	return agent
}
