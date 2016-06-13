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
var defaultFlags *flag.FlagSet

type CommandArgs struct {
	PerformUpgrade bool
	DisplayVersion bool
	DetectOS       bool
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: appcanary [COMMAND] [OPTIONS]\nOptions:\n")

	defaultFlags.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n"+
		"\tupgrade\t\tUpgrade system packages to nearest safe version (Ubuntu only)\n"+
		"\tdetect-os\tDetect current operating system\n")
}

func setFlagset(env *agent.Env, cmdargs *CommandArgs) {
	// httptest, used in client.test, sets a usage flag
	// that leaks when you use the 'global' FlagSet.
	defaultFlags = flag.NewFlagSet("Default", flag.ExitOnError)
	defaultFlags.Usage = usage
	defaultFlags.StringVar(&env.ConfFile, "conf", env.ConfFile, "Set the config file")
	defaultFlags.StringVar(&env.VarFile, "server", env.VarFile, "Set the server file")

	defaultFlags.BoolVar(&env.DryRun, "dry-run", false, "Only print, and do not execute, potentially destructive commands")
	// -version will always override all other args
	defaultFlags.BoolVar(&cmdargs.DisplayVersion, "version", false, "Display version information")

	if !env.Prod {
		defaultFlags.StringVar(&env.BaseUrl, "url", env.BaseUrl, "Set the endpoint")
	}

}

func parseArguments(env *agent.Env, cmdargs *CommandArgs) {
	setFlagset(env, cmdargs)

	switch os.Args[1] {
	case "upgrade":
		cmdargs.PerformUpgrade = true
		defaultFlags.Parse(os.Args[2:])
	case "detect-os":
		// ignore all flags, since we'll just quit
		cmdargs.DetectOS = true
	default:
		defaultFlags.Parse(os.Args[1:])
	}
}

func main() {
	agent.InitEnv(os.Getenv("CANARY_ENV"))
	env := agent.FetchEnv()

	// parse the args
	cmdargs := &CommandArgs{}
	parseArguments(env, cmdargs)

	if cmdargs.DisplayVersion {
		fmt.Println(CanaryVersion)
		os.Exit(0)
	}

	if cmdargs.DetectOS {
		guess, err := detect.DetectOS()
		if err == nil {
			fmt.Printf("%s/%s\n", guess.Distro, guess.Release)
		} else {
			fmt.Printf(err.Error())
		}
		os.Exit(0)
	}

	// let's get started eh
	// start the logger
	agent.InitLogging()
	log := agent.FetchLog()

	fmt.Println(env.Logo)

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

	if cmdargs.PerformUpgrade {
		a.PerformUpgrade()
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
