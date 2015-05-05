package main

import (
	"fmt"
	"os"
	"time"

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

	// we prob can't reliably fingerprint servers.
	// so instead, we assign a uuid by registering
	if a.FirstRun() {

		log.Debug("Found no server config. Let's register!")
		err := a.RegisterServer()

		if err != nil {
			log.Fatal(err)
		}

	}

	// Add hooks to files, and push them over
	// whenever they change
	a.StartWatching()

	// send a heartbeat every ~60min, forever
	go func() {
		tick := time.Tick(env.HeartbeatDuration)

		for {
			err := a.Heartbeat()
			if err != nil {
				log.Fatal("<3 ", err)
			}
			<-tick
		}
	}()

	defer a.CloseWatches()

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
