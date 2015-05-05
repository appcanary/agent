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
	ConfPath          string
	ConfFile          string
	VarPath           string
	VarFile           string
	HeartbeatDuration time.Duration
}

var env = &Umwelten{}

func Init(env_str string) {
	env.Env = env_str
	if env_str != "test" && env_str != "debug" {
		env.Prod = true
	}

	// to be overriden by cli options
	if env.Prod {
		logging.SetLevel(logging.NOTICE, "canary-agent")
		env.BaseUrl = PROD_URL

		env.Logo = PROD_LOGO

		env.ConfPath = DEFAULT_CONF_PATH
		env.ConfFile = DEFAULT_CONF_FILE

		env.VarPath = DEFAULT_VAR_PATH
		env.VarFile = DEFAULT_VAR_FILE

		env.HeartbeatDuration = DEFAULT_HEARTBEAT_DURATION
	} else {
		logging.SetLevel(logging.DEBUG, "canary-agent")

		// filepath.Abs was resolving to a different folder
		// depending on whether it was run from main or a test
		DEV_CONF_PATH, _ = filepath.Abs("test/data")
		if _, err := os.Stat(DEV_CONF_PATH); err != nil {
			DEV_CONF_PATH, _ = filepath.Abs("../test/data")
		}
		DEV_CONF_FILE = filepath.Join(DEV_CONF_PATH, "test2.conf")

		DEV_VAR_PATH, _ = filepath.Abs("test/var")
		if _, err := os.Stat(DEV_VAR_PATH); err != nil {
			DEV_VAR_PATH, _ = filepath.Abs("../test/var")
		}
		DEV_VAR_FILE = filepath.Join(DEV_VAR_PATH, "server.conf")

		env.BaseUrl = DEV_URL

		env.Logo = DEV_LOGO

		env.ConfPath = DEV_CONF_PATH
		env.ConfFile = DEV_CONF_FILE

		env.VarPath = DEV_VAR_PATH
		env.VarFile = DEV_VAR_FILE

		env.HeartbeatDuration = DEV_HEARTBEAT_DURATION

	}
}

func Fetch() *Umwelten {
	return env
}
