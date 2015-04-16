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

	log.Debug(env.ConfPath)
	_, err := os.Stat(env.ConfPath)
	if os.IsNotExist(err) {
		log.Notice("We need to implement getting the env info from the user")
		return
	}

	// slurp env, instantiate agent
	conf := agent.NewConfFromFile(env.ConfPath)
	a := agent.NewAgent(conf)

	// TODO: skip registering if uuid is in conf
	err = a.RegisterServer()
	// realistically, the agent doesn't have to be aware of
	// how we're going to be queueing retries
	if err != nil {
		log.Fatal(err)
	}

	// TODO: LOOP FOREVER
	err = a.Heartbeat()
	if err != nil {
		log.Fatal("<3 ", err)
	}

	// TODO: submit watched files
	err = a.RegisterApps()

	if err != nil {
		log.Fatal("RegisterApps ", err)
	}

	defer a.CloseWatches()

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
