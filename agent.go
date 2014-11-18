package agent

import (
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"

	"github.com/mveytsman/canary-agent/data"
)

var lg = logging.MustGetLogger("app-canary")

type Agent struct {
	conf *Conf
	db   gorm.DB
}

func NewAgent(confPath string) *Agent {
	agent := &Agent{}

	agent.conf = NewConfFromFile(confPath)
	agent.db = data.Initialize()

	// load the existing gemfiles yo'

	agent.NewWatchedFile(agent.conf.Ruby.Projects[0][0], agent.conf.Ruby.Projects[0][1]+"/Gemfile.lock")

	return agent
}
