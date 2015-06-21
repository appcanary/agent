package umwelten

import (
	"os"
	"path/filepath"
	"time"

	"github.com/op/go-logging"
)

var Log = logging.MustGetLogger("canary-agent")

type Umwelten struct {
	Logo              string
	Env               string
	Prod              bool
	BaseUrl           string
	ConfFile          string
	VarFile           string
	HeartbeatDuration time.Duration
	LogFile           *os.File
}

var env = &Umwelten{}

func Init(env_str string) {
	env.Env = env_str
	if env_str != "test" && env_str != "debug" {
		env.Prod = true
	}

	stdoutBackend := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), logging.GlogFormatter)

	// to be overriden by cli options
	if env.Prod {
		logging.SetLevel(logging.INFO, "canary-agent")
		env.BaseUrl = PROD_URL

		env.Logo = PROD_LOGO

		env.ConfFile = DEFAULT_CONF_FILE

		env.VarFile = DEFAULT_VAR_FILE

		env.HeartbeatDuration = DEFAULT_HEARTBEAT_DURATION

		//TODO: This needs to happen outside of the init, because the init is called before we parse command line flags and we eventually want the log file location to be user secified.
		//I think the best thing to do is to refactor this bit of code and how we handle dev/prod mode to work better with the flags package
		var err error
		env.LogFile, err = os.OpenFile(DEFAULT_LOG_FILE, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			Log.Error("Can't open log file", err) //INCEPTION
			os.Exit(1)
		} else {
			fileBackend := logging.NewBackendFormatter(logging.NewLogBackend(env.LogFile, "", 0), logging.GlogFormatter)
			logging.SetBackend(fileBackend, stdoutBackend)
		}

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

		// set dev vals

		env.BaseUrl = DEV_URL

		env.Logo = DEV_LOGO

		env.ConfFile = DEV_CONF_FILE

		env.VarFile = DEV_VAR_FILE

		env.HeartbeatDuration = DEV_HEARTBEAT_DURATION

		logging.SetLevel(logging.DEBUG, "canary-agent")
		logging.SetBackend(stdoutBackend)

	}
}

func Fetch() *Umwelten {
	return env
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
