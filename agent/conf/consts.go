package conf

import "time"

// consts can't be outputs of functions?
var DEV_CONF_PATH string
var DEV_CONF_FILE string

var DEV_VAR_PATH string
var DEV_VAR_FILE string

// env vars
const (
	PROD_URL = "https://www.appcanary.com"
	DEV_URL  = "http://localhost:4000"

	DEFAULT_CONF_PATH      = "/etc/appcanary/"
	DEFAULT_CONF_FILE_BASE = DEFAULT_CONF_PATH + "agent"
	DEFAULT_VAR_PATH       = "/var/db/appcanary/"
	DEFAULT_VAR_FILE_BASE  = DEFAULT_VAR_PATH + "server"
	DEFAULT_CONF_FILE      = DEFAULT_CONF_FILE_BASE + ".yml"
	DEFAULT_VAR_FILE       = DEFAULT_VAR_FILE_BASE + ".yml"
	OLD_DEFAULT_CONF_FILE  = DEFAULT_CONF_FILE_BASE + ".conf"
	OLD_DEFAULT_VAR_FILE   = DEFAULT_VAR_FILE_BASE + ".conf"

	DEFAULT_HEARTBEAT_DURATION = 1 * time.Hour
	DEV_HEARTBEAT_DURATION     = 10 * time.Second

	DEFAULT_SYNC_ALL_DURATION = 24 * time.Hour
	DEV_SYNC_ALL_DURATION     = 30 * time.Second

	DEFAULT_LOG_FILE = "/var/log/appcanary.log"
)

// api endpoints
const (
	API_VERSION   = "/api/v1/agent/"
	API_HEARTBEAT = API_VERSION + "heartbeat"
	API_SERVERS   = API_VERSION + "servers"
)

// file polling
const (
	DEFAULT_POLL_SLEEP = 5 * time.Minute
	// test env.PollSleep is 1second
	// test poll sleep is double to give the fs time to flush
	DEV_POLL_SLEEP  = time.Second
	TEST_POLL_SLEEP = (time.Second + (150 * time.Millisecond)) * 2
)

// trolol
const (
	DEV_LOGO = `
        ********** ********  ******** **********
       /////**/// /**/////  **////// /////**///
           /**    /**      /**           /**
           /**    /******* /*********    /**
           /**    /**////  ////////**    /**
           /**    /**             /**    /**
           /**    /******** ********     /**
           //     //////// ////////      //
`
	PROD_LOGO = `


     __     _____   _____     ___     __      ___      __     _ __   __  __
   /'__` + "`" + `\  /\ '__` + "`" + `\/\ '__` + "`" + `\  /'___\ /'__` + "`" + `\  /' _ ` + "`" + `\  /'__` + "`" + `\  /\` + "`" + `'__\/\ \/\ \
  /\ \L\.\_\ \ \L\ \ \ \L\ \/\ \__//\ \L\.\_/\ \/\ \/\ \L\.\_\ \ \/ \ \ \_\ \
  \ \__/.\_\\ \ ,__/\ \ ,__/\ \____\ \__/.\_\ \_\ \_\ \__/.\_\\ \_\  \/` + "`" + `____ \
   \/__/\/_/ \ \ \/  \ \ \/  \/____/\/__/\/_/\/_/\/_/\/__/\/_/ \/_/   ` + "`" + `/___/> \
              \ \_\   \ \_\                                              /\___/
               \/_/    \/_/                                              \/__/


`
)
