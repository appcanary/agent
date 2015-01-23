package main

import (
	"os"

	"github.com/stateio/canary-agent/agent"
)

func main() {
	done := make(chan os.Signal, 1)
	conf := agent.NewConfFromFile("agent/testdata/test2.conf")
	a := agent.NewAgent(conf)
	defer a.CloseWatches()

	//	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
