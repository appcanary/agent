package agent

import (
	"os"
	"path/filepath"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("canary-agent")

type Env struct {
	Logo                       string
	Env                        string
	Prod                       bool
	BaseUrl                    string
	ConfFile                   string
	VarFile                    string
	HeartbeatDuration          time.Duration
	LogFile                    *os.File
	DebianLikeDistributionFile string
}

var env = &Env{}

func FetchEnv() *Env {
	return env
}

func FetchLog() *logging.Logger {
	return log
}

func InitEnv(env_str string) {
	env.Env = env_str
	if env_str != "test" && env_str != "debug" {
		env.Prod = true
	}

	// to be overriden by cli options
	if env.Prod {

		env.BaseUrl = PROD_URL

		env.Logo = PROD_LOGO

		env.ConfFile = DEFAULT_CONF_FILE

		env.VarFile = DEFAULT_VAR_FILE

		env.HeartbeatDuration = DEFAULT_HEARTBEAT_DURATION

		env.DebianLikeDistributionFile = DEFAULT_DEBIAN_LIKE_DISTRIBUTION_FILE

	} else {
		// ###### resolve path
		// filepath.Abs was resolving to a different folder
		// depending on whether it was run from main or a test
		DEV_CONF_PATH, _ = filepath.Abs("test/data")
		if _, err := os.Stat(DEV_CONF_PATH); err != nil {
			DEV_CONF_PATH, _ = filepath.Abs("../test/data")
		}
		DEV_CONF_FILE = filepath.Join(DEV_CONF_PATH, "test.conf")

		DEV_VAR_PATH, _ = filepath.Abs("test/var")
		if _, err := os.Stat(DEV_VAR_PATH); err != nil {
			DEV_VAR_PATH, _ = filepath.Abs("../test/var")
		}
		DEV_VAR_FILE = filepath.Join(DEV_VAR_PATH, "server.conf")

		DEV_DEBIAN_LIKE_DISTRIBUTION_FILE, _ := filepath.Abs("../test/data/issue")

		// set dev vals

		env.BaseUrl = DEV_URL

		env.Logo = DEV_LOGO

		env.ConfFile = DEV_CONF_FILE

		env.VarFile = DEV_VAR_FILE

		env.HeartbeatDuration = DEV_HEARTBEAT_DURATION

		env.DebianLikeDistributionFile = DEV_DEBIAN_LIKE_DISTRIBUTION_FILE

	}
}

func InitLogging() {
	stdoutBackend := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), logging.GlogFormatter)
	var err error
	if env.Prod {
		logging.SetLevel(logging.INFO, "canary-agent")

		conf := NewConfFromEnv()
		var logPath string
		if conf.LogPath != "" {
			logPath = conf.LogPath
		} else {
			logPath = DEFAULT_LOG_FILE
		}

		env.LogFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Can't open log file", err) //INCEPTION
			os.Exit(1)
		} else {
			fileBackend := logging.NewBackendFormatter(logging.NewLogBackend(env.LogFile, "", 0), logging.GlogFormatter)
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

func ApiPath(resource string) string {
	return env.BaseUrl + resource
}
