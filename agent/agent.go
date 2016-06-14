package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var CanaryVersion string

type Agent struct {
	conf   *Conf
	client Client
	server *Server
	files  Watchers
}

func NewAgent(version string, conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, files: Watchers{}}

	// Find out what we need about machine
	// Fills out server conf if some values are missing
	agent.server = NewServer(conf, conf.ServerConf)

	if len(clients) > 0 {
		agent.client = clients[0]
	} else {
		agent.client = NewClient(conf.ApiKey, agent.server)
	}

	CanaryVersion = version
	return agent
}

// instantiate structs, fs hook
func (agent *Agent) StartWatching() {
	for _, f := range agent.conf.Files {
		var watcher Watcher

		if f.Process == "" {
			watcher = NewFileWatcherWithHook(f.Path, agent.OnChange)
		} else {
			watcher = NewProcessWatcherWithHook(f.Process, agent.OnChange)
		}

		agent.files = append(agent.files, watcher)
	}
}

func (agent *Agent) OnChange(file Watcher) {
	log.Infof("File change: %s", file.Path())

	// should probably be in the actual hook code
	contents, err := file.Contents()

	if err != nil {
		// we couldn't read it; something weird is happening
		// let's just wait until this callback gets issued
		// again when the file reappears.
		log.Infof("File contents error: %s", err)
		return
	}
	err = agent.client.SendFile(file.Path(), file.Kind(), contents)
	if err != nil {
		// TODO: some kind of queuing mechanism to keep trying
		// beyond the exponential backoff in the client.
		// What if the connection fails for whatever reason?
		log.Infof("Sendfile error: %s", err)
	}
}

func (agent *Agent) SyncAllFiles() {
	log.Info("Synching all files.")

	for _, f := range agent.files {
		agent.OnChange(f)
	}
}

func (agent *Agent) Heartbeat() error {
	return agent.client.Heartbeat(agent.server.UUID, agent.files)
}

func (agent *Agent) FirstRun() bool {
	// the configuration didn't find a server uuid
	return agent.server.IsNew()
}

func (agent *Agent) RegisterServer() error {
	uuid, err := agent.client.CreateServer(agent.server)

	if err != nil {
		return err
	}
	agent.server.UUID = uuid
	agent.conf.ServerConf.UUID = uuid
	agent.conf.Save()
	return nil
}

func (agent *Agent) PerformUpgrade() error {
	package_list, err := agent.client.FetchUpgradeablePackages()

	if err != nil {
		log.Fatalf("Can't fetch upgrade info: %s", err)
	}

	if agent.server.DebianLike() {
		agent.runDebianUpgrade(package_list)
	} else {
		log.Fatal("Sorry, we don't support your operating system at the moment. Is this a mistake? Run `appcanary detect-os` and tell us about it at support@appcanary.com")
	}
	return nil
}

func (agent *Agent) runDebianUpgrade(package_list map[string]string) {
	updateCmd := "apt-get"
	updateArg := []string{"update", "-q"}

	installCmd := "apt-get"
	installArg := []string{"install", "--only-upgrade", "--no-install-recommends", "-y", "-q"}

	for name, version := range package_list {
		installArg = append(installArg, name+"="+version)
	}

	if env.DryRun {
		log.Info("Running upgrade in dry-run mode...")
	}

	err := agent.runCmd(updateCmd, updateArg)
	if err != nil {
		log.Fatal(err)
	}

	err = agent.runCmd(installCmd, installArg)
	if err != nil {
		log.Fatal(err)
	}
}

func (agent *Agent) runCmd(cmd_name string, args []string) error {
	_, err := exec.LookPath(cmd_name)

	if err != nil {
		log.Fatal("Can't find " + cmd_name)
	}

	cmd := exec.Command(cmd_name, args...)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	log.Infof("Running: %s %s", cmd_name, strings.Join(args, " "))
	if !env.DryRun {
		if err := cmd.Start(); err != nil {
			log.Fatalf("Was unable to start %s. Error: %v", cmd_name, err)
		}

		err = cmd.Wait()
		fmt.Println(string(output.Bytes()))

		return err
	} else {
		return nil
	}

}

// This has to be called before exiting
func (agent *Agent) CloseWatches() {
	for _, file := range agent.files {
		file.Stop()
	}
}
