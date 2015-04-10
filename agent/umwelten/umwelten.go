package umwelten

type Umwelten struct {
	Logo     string
	Env      string
	Prod     bool
	BaseUrl  string
	ConfPath string
}

var env = &Umwelten{}

func Init(env_str string) {
	env.Env = env_str
	if env_str != "test" && env_str != "debug" {
		env.Prod = true
	}

	if env.Prod {
		env.BaseUrl = PROD_URL
		env.Logo = PROD_LOGO
		env.ConfPath = DEFAULT_CONF_PATH
	} else {
		env.BaseUrl = DEV_URL
		env.Logo = DEV_LOGO
		env.ConfPath = DEV_CONF_PATH
	}
}

func Fetch() *Umwelten {
	return env
}
