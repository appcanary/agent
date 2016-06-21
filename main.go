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

type CommandToPerform int

const (
	PerformAgentLoop CommandToPerform = iota
	PerformUpgrade
	PerformDisplayVersion
	PerformDetectOS
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: appcanary [COMMAND] [OPTIONS]\nOptions:\n")

	defaultFlags.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n"+
		"\t[none]\t\tStart the agent\n"+
		"\tupgrade\t\tUpgrade system packages to nearest safe version (Ubuntu only)\n"+
		"\tdetect-os\tDetect current operating system\n")
}

func parseFlags(argRange int, env *agent.Env, performCmd *CommandToPerform) {
	var displayVersionFlagged bool
	// httptest, used in client.test, sets a usage flag
	// that leaks when you use the 'global' FlagSet.
	defaultFlags = flag.NewFlagSet("Default", flag.ExitOnError)
	defaultFlags.Usage = usage
	defaultFlags.StringVar(&env.ConfFile, "conf", env.ConfFile, "Set the config file")
	defaultFlags.StringVar(&env.VarFile, "server", env.VarFile, "Set the server file")

	defaultFlags.BoolVar(&env.DryRun, "dry-run", false, "Only print, and do not execute, potentially destructive commands")
	// -version will always override all other args
	defaultFlags.BoolVar(&displayVersionFlagged, "version", false, "Display version information")

	if !env.Prod {
		defaultFlags.StringVar(&env.BaseUrl, "url", env.BaseUrl, "Set the endpoint")
	}

	defaultFlags.Parse(os.Args[argRange:])

	if displayVersionFlagged {
		*performCmd = PerformDisplayVersion
	}
}

func parseArguments(env *agent.Env) CommandToPerform {
	var performCmd CommandToPerform

	if len(os.Args) < 2 {
		return PerformAgentLoop
	}

	argRange := 1
	// TODO: replace this boolean switch statement
	// with some kind of enum dispatch
	switch os.Args[1] {
	case "upgrade":
		argRange = 2
		// defaultFlags.Parse(os.Args[2:])
		performCmd = PerformUpgrade
	case "detect-os":
		argRange = 2
		performCmd = PerformDetectOS
	default:
		// defaultFlags.Parse(os.Args[1:])
		performCmd = PerformAgentLoop
	}

	parseFlags(argRange, env, &performCmd)
	return performCmd
}

func runDisplayVersion() {
	fmt.Println(CanaryVersion)
	os.Exit(0)
}

func runDetectOS() {
	guess, err := detect.DetectOS()
	if err == nil {
		fmt.Printf("%s/%s\n", guess.Distro, guess.Release)
	} else {
		fmt.Printf("%v\n", err.Error())
	}
	os.Exit(0)
}

func initialize(env *agent.Env) *agent.Agent {
	// let's get started eh
	// start the logger
	agent.InitLogging()
	log := agent.FetchLog()

	fmt.Println(env.Logo)

	// slurp env, instantiate agent
	conf := agent.NewConfFromEnv()

	if conf.ApiKey == "" {
		log.Fatal("There's no API key set. Get yours from https://appcanary.com/settings and set it in /etc/appcanary/agent.conf")
	}

	a := agent.NewAgent(CanaryVersion, conf)
	a.DoneChannel = make(chan os.Signal, 1)

	// we prob can't reliably fingerprint servers.
	// so instead, we assign a uuid by registering
	if a.FirstRun() {
		log.Debug("Found no server config. Let's register!")

		for err := a.RegisterServer(); err != nil; {
			// we don't need to wait here because of the backoff
			// exponential decay library; by the time we hit this
			// point we've been trying for about, what, an hour?
			log.Infof("Register server error: %s", err)
			err = a.RegisterServer()
		}

	}

	// Now that we're registered,
	// let's init our watchers. We auto sync on watcher init.
	a.BuildAndSyncWatchers()
	return a
}

func runUpgrade(a *agent.Agent) {
	a.PerformUpgrade()
	os.Exit(0)
}

func runAgentLoop(env *agent.Env, a *agent.Agent) {
	log := agent.FetchLog()
	// Add hooks to files, and push them over
	// whenever they change
	a.StartPolling()

	// send a heartbeat every ~60min, forever
	go func() {
		tick := time.Tick(env.HeartbeatDuration)

		for {
			err := a.Heartbeat()
			if err != nil {
				log.Infof("<3 error: %s", err)
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

	// wait for the right signal?
	// signal.Notify(done, os.Interrupt, os.Kill)

	// block forever
	<-a.DoneChannel
}

func main() {
	agent.InitEnv(os.Getenv("CANARY_ENV"))
	env := agent.FetchEnv()

	// parse the args
	switch parseArguments(env) {

	case PerformDisplayVersion:
		runDisplayVersion()

	case PerformDetectOS:
		runDetectOS()

	case PerformUpgrade:
		a := initialize(env)
		runUpgrade(a)

	case PerformAgentLoop:
		a := initialize(env)
		runAgentLoop(env, a)
	}

	// Close the logfile when we exit
	if env.LogFileHandle != nil {
		defer env.LogFileHandle.Close()
	}

}
