package main

import (
	"fmt"
	"os"

	"github.com/stateio/canary-agent/agent"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var env = umwelten.Fetch()

func main() {
	done := make(chan os.Signal, 1)

	umwelten.Init(os.Getenv("CANARY_ENV"))

	fmt.Println(env.Logo)

	_, err := os.Stat(env.ConfPath)
	if os.IsNotExist(err) {
		fmt.Println("We need to implement getting the env info from the user")
		return
	}

	// slurp env, instantiate agent
	conf := agent.NewConfFromFile(env.ConfPath)
	a := agent.NewAgent(conf)
	defer a.CloseWatches()

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
