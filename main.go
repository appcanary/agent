package main

import (
	"fmt"
	"os"

	"github.com/stateio/canary-agent/agent"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var env = umwelten.Fetch()
var log = umwelten.Log

func main() {
	done := make(chan os.Signal, 1)

	umwelten.Init(os.Getenv("CANARY_ENV"))

	fmt.Println(env.Logo)

	// slurp env, instantiate agent
	conf := agent.NewConfFromEnv()
	a := agent.NewAgent(conf)

	// we probably can't reliably fingerprint servers, so
	// we don't even try. therefore, if we find an existing
	// uuid, we should reuse it. TODO: what if the uuid fails
	// during the heartbeat?
	if a.FirstRun() {

		log.Debug("Found no server config. Let's register!")
		err := a.RegisterServer()

		// the agent doesn't have to be aware of
		// how we're going to be queueing retries
		if err != nil {
			log.Fatal(err)
		}

	}

	// TODO: LOOP FOREVER
	err := a.Heartbeat()
	if err != nil {
		log.Fatal("<3 ", err)
	}

	defer a.CloseWatches()

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
