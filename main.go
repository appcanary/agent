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
	PerformProcessInspection
	PerformProcessInspectionJsonDump
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: appcanary [COMMAND] [OPTIONS]\nOptions:\n")

	defaultFlags.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n"+
		"\t[none]\t\t\tStart the agent\n"+
		"\tupgrade\t\t\tUpgrade system packages to nearest safe version (Ubuntu only)\n"+
		"\tinspect-processes\tSend your process library information to Appcanary\n"+
		"\tdetect-os\t\tDetect current operating system\n")
}

func parseFlags(argRange int, env *agent.Env) {
	var displayVersionFlagged bool
	// httptest, used in client.test, sets a usage flag
	// that leaks when you use the 'global' FlagSet.
	defaultFlags = flag.NewFlagSet("Default", flag.ExitOnError)
	defaultFlags.Usage = usage
	defaultFlags.StringVar(&env.ConfFile, "conf", env.ConfFile, "Set the config file")
	defaultFlags.StringVar(&env.VarFile, "server", env.VarFile, "Set the server file")

	defaultFlags.BoolVar(&env.DryRun, "dry-run", false, "Only print, and do not execute, potentially destructive commands")
	// -version is handled in parseArguments, but is set here for the usage print out
	defaultFlags.BoolVar(&displayVersionFlagged, "version", false, "Display version information")

	defaultFlags.BoolVar(&env.FailOnConflict, "fail-on-conflict", false, "Should upgrade encounter a conflict with configuration files, abort (default: old configuration files are kept, or updated if not modified)")

	if !env.Prod {
		defaultFlags.StringVar(&env.BaseUrl, "url", env.BaseUrl, "Set the endpoint")
	}

	defaultFlags.Parse(os.Args[argRange:])
}

func parseArguments(env *agent.Env) CommandToPerform {
	var performCmd CommandToPerform

	if len(os.Args) < 2 {
		return PerformAgentLoop
	}

	// if first arg is a command,
	// flags will follow in os.Args[2:]
	// else in os.Args[1:]
	argRange := 2
	switch os.Args[1] {
	case "upgrade":
		performCmd = PerformUpgrade
	case "detect-os":
		performCmd = PerformDetectOS
	case "inspect-processes":
		performCmd = PerformProcessInspection
	case "inspect-processes-json":
		performCmd = PerformProcessInspectionJsonDump
	case "-version":
		performCmd = PerformDisplayVersion
	case "--version":
		performCmd = PerformDisplayVersion
	default:
		argRange = 1
		performCmd = PerformAgentLoop
	}

	parseFlags(argRange, env)
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

	// If the config sets a startup delay, we wait to boot up here
	if conf.StartupDelay != 0 {
		delay := time.Duration(conf.StartupDelay) * time.Second
		tick := time.Tick(delay)
		<-tick
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

func runProcessInspection(a *agent.Agent) {
	agent.ShipProcessMap(a)
	fmt.Println("Done! Check https://appcanary.com/")
	os.Exit(0)
}

func runProcessInspectionDump() {
	agent.DumpProcessMap()
	os.Exit(0)
}

func runUpgrade(a *agent.Agent) {
	log := agent.FetchLog()
	log.Info("Running upgrade...")
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

func checkYourPrivilege() {
	if os.Getuid() != 0 && os.Geteuid() != 0 {
		fmt.Println("Cannot run unprivileged - must be root (UID=0)")
		os.Exit(13)
	}
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

	case PerformProcessInspection:
		checkYourPrivilege()
		a := initialize(env)
		runProcessInspection(a)

	case PerformProcessInspectionJsonDump:
		checkYourPrivilege()
		agent.InitLogging()
		runProcessInspectionDump()

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
