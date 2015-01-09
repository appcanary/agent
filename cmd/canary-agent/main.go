package main

import (
	"os"
	"os/signal"

	"github.com/mveytsman/canary-agent"
)

func main() {
	done := make(chan os.Signal, 1)
	conf := agent.NewConfFromFile("testdata/test.conf")
	agent := agent.NewAgent(conf)
	defer agent.CloseWatches()

	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
