package agent

import "os"

var CanaryVersion string

type Agent struct {
	conf        *Conf
	client      Client
	server      *Server
	files       Watchers
	DoneChannel chan os.Signal
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
func (agent *Agent) StartPolling() {
	for _, watcher := range agent.files {
		watcher.Start()
	}
}

func (agent *Agent) BuildAndSyncWatchers() {
	for _, f := range agent.conf.Files {
		var watcher Watcher

		if f.Process != "" {
			watcher = NewProcessWatcher(f.Process, agent.OnChange)
		} else if f.Command != "" {
			watcher = NewCommandOutputWatcher(f.Command, agent.OnChange)
		} else if f.Path != "" {
			watcher = NewFileWatcher(f.Path, agent.OnChange)
		}
		agent.files = append(agent.files, watcher)
	}
	agent.files = append(agent.files, NewAllProcessWatcher(agent.OnChange))
}

func (agent *Agent) OnChange(w Watcher) {
	switch wt := w.(type) {
	default:
		log.Errorf("Don't know what to do with %T", wt)
	case TextWatcher:
		agent.handleTextChange(wt)
	case ProcessWatcher:
		agent.handleProcessChange(wt)
	}
}

func (agent *Agent) handleProcessChange(pw ProcessWatcher) {
	match := pw.Match()
	if match == "*" {
		log.Infof("Shipping process map")
	} else {
		log.Infof("Shipping process map for %s", match)
	}
	agent.client.SendProcessState(match, pw.StateJson())
}

func (agent *Agent) handleTextChange(tw TextWatcher) {
	log.Infof("File change: %s", tw.Path())

	// should probably be in the actual hook code
	contents, err := tw.Contents()
	if err != nil {
		// we couldn't read it; something weird is happening let's just wait
		// until this callback gets issued again when the file reappears.
		log.Infof("File contents error: %s", err)
		return
	}

	err = agent.client.SendFile(tw.Path(), tw.Kind(), contents)
	if err != nil {
		// TODO: some kind of queuing mechanism to keep trying beyond the
		// exponential backoff in the client. What if the connection fails for
		// whatever reason?
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

func (agent *Agent) PerformUpgrade() {
	var cmds UpgradeSequence
	packageList, err := agent.client.FetchUpgradeablePackages()

	if err != nil {
		log.Fatalf("Can't fetch upgrade info: %s", err)
	}

	if len(packageList) == 0 {
		log.Info("No vulnerable packages reported. Carry on!")
		return
	}

	if agent.server.IsUbuntu() {
		cmds = buildDebianUpgrade(packageList)
	} else if agent.server.IsCentOS() {
		cmds = buildCentOSUpgrade(packageList)
	} else {
		log.Fatal("Sorry, we don't support your operating system at the moment. Is this a mistake? Run `appcanary detect-os` and tell us about it at support@appcanary.com")
	}

	err = executeUpgradeSequence(cmds)
	if err != nil {
		log.Fatal(err)
	}
}

// This has to be called before exiting
func (agent *Agent) CloseWatches() {
	for _, file := range agent.files {
		file.Stop()
	}
}
