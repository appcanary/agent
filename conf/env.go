package conf

import (
	"os"
	"path/filepath"
	"time"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("canary-agent")

type Env struct {
	Env               string
	Prod              bool
	DryRun            bool
	FailOnConflict    bool
	Logo              string
	BaseUrl           string
	ConfFile          string
	VarFile           string
	LogFile           string
	LogFileHandle     *os.File
	HeartbeatDuration time.Duration
	SyncAllDuration   time.Duration
	PollSleep         time.Duration
}

var env = &Env{
	Prod:              true,
	DryRun:            false,
	FailOnConflict:    false,
	Logo:              PROD_LOGO,
	BaseUrl:           PROD_URL,
	ConfFile:          DEFAULT_CONF_FILE,
	VarFile:           DEFAULT_VAR_FILE,
	LogFile:           DEFAULT_LOG_FILE,
	HeartbeatDuration: DEFAULT_HEARTBEAT_DURATION,
	SyncAllDuration:   DEFAULT_SYNC_ALL_DURATION,
	PollSleep:         DEFAULT_POLL_SLEEP}

func FetchEnv() *Env {
	return env
}

func FetchLog() *logging.Logger {
	return log
}

func InitEnv(envStr string) {
	env.Env = envStr
	if envStr == "test" || envStr == "debug" {
		env.Prod = false
	}

	// to be overriden by cli options
	if env.Prod == false {
		// ###### resolve path
		// filepath.Abs was resolving to a different folder
		// depending on whether it was run from main or a test
		DEV_CONF_PATH, _ = filepath.Abs("test/data")

		if _, err := os.Stat(DEV_CONF_PATH); err != nil {
			DEV_CONF_PATH, _ = filepath.Abs("../test/data")
		}

		DEV_CONF_FILE = filepath.Join(DEV_CONF_PATH, "agent.yml")
		OLD_DEV_CONF_FILE = filepath.Join(DEV_CONF_PATH, "old_toml_test.conf")

		DEV_VAR_FILE = filepath.Join(DEV_CONF_PATH, "server.yml")
		OLD_DEV_VAR_FILE = filepath.Join(DEV_CONF_PATH, "old_toml_server.conf")

		// set dev vals

		env.BaseUrl = DEV_URL

		env.Logo = DEV_LOGO

		env.ConfFile = DEV_CONF_FILE

		env.VarFile = DEV_VAR_FILE

		env.HeartbeatDuration = DEV_HEARTBEAT_DURATION
		env.SyncAllDuration = DEV_SYNC_ALL_DURATION

		env.PollSleep = DEV_POLL_SLEEP

	}
}

func InitLogging() {
	// TODO: SetLevel must come before SetBackend
	format := logging.MustStringFormatter("%{time} %{pid} %{shortfile}] %{message}")
	stdoutBackend := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format)
	if env.Prod {
		logging.SetLevel(logging.INFO, "canary-agent")

		conf, err := NewConfFromEnv()
		if err != nil {
			log.Fatal(err)
		}

		var logPath string
		if conf.LogPath != "" {
			logPath = conf.LogPath
		} else {
			logPath = env.LogFile
		}

		env.LogFileHandle, err = os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Can't open log file", err) //INCEPTION
			os.Exit(1)
		} else {
			fileBackend := logging.NewBackendFormatter(logging.NewLogBackend(env.LogFileHandle, "", 0), logging.GlogFormatter)
			logging.SetBackend(fileBackend, stdoutBackend)
		}
	} else {
		logging.SetLevel(logging.DEBUG, "canary-agent")
		logging.SetBackend(stdoutBackend)
	}
}

func ApiHeartbeatPath(ident string) string {
	return ApiPath(API_HEARTBEAT) + "/" + ident
}

func ApiServersPath() string {
	return ApiPath(API_SERVERS)
}

func ApiServerPath(ident string) string {
	return ApiServersPath() + "/" + ident
}

func ApiServerProcsPath(ident string) string {
	return ApiServerPath(ident) + "/processes"
}

func ApiPath(resource string) string {
	return env.BaseUrl + resource
}
