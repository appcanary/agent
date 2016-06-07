package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/appcanary/agent/agent"
	"github.com/appcanary/agent/agent/detect"
)

var CanaryVersion string
var flagset *flag.FlagSet

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: appcanary [OPTION]\n")
	flagset.PrintDefaults()
}

func setAndPrintFlags(env *agent.Env) {
	var flaggedVersion, flaggedDetectOS *bool
	// httptest, used in client.test, sets a usage flag
	// that leaks when you use the 'global' FlagSet.
	flagset = flag.NewFlagSet("Default", flag.ExitOnError)
	flagset.Usage = usage

	flagset.StringVar(&env.ConfFile, "conf", env.ConfFile, "Set the config file")
	flagset.StringVar(&env.VarFile, "server", env.VarFile, "Set the server file")
	flagset.BoolVar(&env.PerformUpgrade, "upgrade", false, "Perform security upgrades")
	flaggedDetectOS = flagset.Bool("detect-os", false, "Guess my operating system")

	if !env.Prod {
		flagset.StringVar(&env.BaseUrl, "url", env.BaseUrl, "Set the endpoint")
	}

	flaggedVersion = flagset.Bool("version", false, "Display version information")
	flagset.Parse(os.Args[1:])

	if flaggedVersion != nil {
		if *flaggedVersion {
			fmt.Println(CanaryVersion)
			os.Exit(0)
		}
	}

	if flaggedDetectOS != nil {
		if *flaggedDetectOS {
			guess, err := detect.DetectOS()
			if err == nil {
				fmt.Printf("%s/%s\n", guess.Distro, guess.Release)
			} else {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
	}
}

func main() {
	agent.InitEnv(os.Getenv("CANARY_ENV"))
	env := agent.FetchEnv()

	setAndPrintFlags(env)

	//start the logger
	fmt.Println(env.Logo)
	agent.InitLogging()
	log := agent.FetchLog()

	done := make(chan os.Signal, 1)

	// slurp env, instantiate agent
	conf := agent.NewConfFromEnv()

	if conf.ApiKey == "" {
		log.Fatal("There's no API key set. Get yours from https://appcanary.com/settings and set it in /etc/appcanary/agent.conf")
	}

	a := agent.NewAgent(CanaryVersion, conf)

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

	if env.PerformUpgrade {
		a.PerformUpgrade()
		// ideally should know to send over update first.
		os.Exit(0)
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

	go func() {
		tick := time.Tick(env.SyncAllDuration)

		for {
			<-tick
			a.SyncAllFiles()
		}
	}()

	defer a.CloseWatches()

	// Close the logfile when we exit
	if env.LogFileHandle != nil {
		defer env.LogFileHandle.Close()
	}

	// wait for the right signal
	// signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
