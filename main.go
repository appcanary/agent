package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stateio/canary-agent/agent"
	"github.com/stateio/canary-agent/agent/umwelten"
)

const CanaryVersion = "0.1 Alpha"

var env = umwelten.Fetch()
var log = umwelten.Log

func usage() {
	fmt.Fprintf(os.Stderr, "canary-agent: agent for https://www.appcanry.com.\nUsage:\n")
	flag.PrintDefaults()
}

func main() {
	umwelten.Init(os.Getenv("CANARY_ENV"))

	flag.Usage = usage

	flag.StringVar(&env.ConfFile, "conf", env.ConfFile, "Set the config file")
	flag.StringVar(&env.VarFile, "server", env.VarFile, "Set the server file")
	version := flag.Bool("version", false, "Display version information")
	flag.Parse()

	if *version {
		fmt.Println(CanaryVersion)
		os.Exit(0)
	}
	done := make(chan os.Signal, 1)

	fmt.Println(env.Logo)

	// slurp env, instantiate agent
	conf := agent.NewConfFromEnv()
	a := agent.NewAgent(conf)

	// we prob can't reliably fingerprint servers.
	// so instead, we assign a uuid by registering
	if a.FirstRun() {
		log.Debug("Found no server config. Let's register!")

		for err := a.RegisterServer(); err != nil; {
			// we don't need to wait here because of the backoff
			// exponential decay library; by the time we hit this
			// point we've been trying for about, what, an hour?
			log.Info("Register server error: %s", err)
			err = a.RegisterServer()
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
				log.Info("<3 error: %s", err)
			}
			<-tick
		}
	}()

	defer a.CloseWatches()

	// Close the logfile when we exit
	if env.LogFile != nil {
		defer env.LogFile.Close()
	}

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
