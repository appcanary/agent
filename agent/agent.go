package agent

var CanaryVersion string

type Agent struct {
	conf   *Conf
	client Client
	server *Server
	files  WatchedFiles
}

func NewAgent(version string, conf *Conf, clients ...Client) *Agent {
	agent := &Agent{conf: conf, files: WatchedFiles{}}

	// Find out what we need about machine
	// Fills out server conf if some values are missing
	agent.server = NewServer(conf.ServerConf)
	agent.conf.Save()

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
		agent.files = append(agent.files, NewWatchedFileWithHook(f.Path, agent.OnFileChange))
	}
}

func (agent *Agent) OnFileChange(file *WatchedFile) {
	log.Info("File change: %s", file.Path)

	// should probably be in the actual hook code
	contents, err := file.Contents()

	if err != nil {
		// we couldn't read it; something weird is happening
		// let's just wait until this callback gets issued
		// again when the file reappears.
		log.Info("File contents error: %s", err)
		return
	}
	err = agent.client.SendFile(file.Path, file.Kind, contents)
	if err != nil {
		// TODO: some kind of queuing mechanism to keep trying
		// beyond the exponential backoff in the client.
		// What if the connection fails for whatever reason?
		log.Info("Sendfile error: %s", err)
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

// This has to be called before exiting
func (agent *Agent) CloseWatches() {
	for _, file := range agent.files {
		file.StopListening()
	}
}
