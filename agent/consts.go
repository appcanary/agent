package agent

import "time"

// consts can't be outputs of functions?
var DEV_CONF_PATH string
var DEV_CONF_FILE string

var DEV_VAR_PATH string
var DEV_VAR_FILE string

// env vars
const (
	PROD_URL = "https://www.appcanary.com"
	DEV_URL  = "http://localhost:3000"

	DEFAULT_CONF_PATH = "/etc/appcanary/"
	DEFAULT_CONF_FILE = DEFAULT_CONF_PATH + "agent.conf"
	DEFAULT_VAR_PATH  = "/var/db/appcanary/"
	DEFAULT_VAR_FILE  = DEFAULT_VAR_PATH + "server.conf"

	DEFAULT_HEARTBEAT_DURATION = 1 * time.Hour
	DEV_HEARTBEAT_DURATION     = 10 * time.Second

	DEFAULT_LOG_FILE = "/var/log/appcanary/agent.log"
)

// api endpoints
const (
	API_VERSION   = "/api/v1/agent/"
	API_HEARTBEAT = API_VERSION + "heartbeat"
	API_SERVERS   = API_VERSION + "servers"
)

// file polling
const (
	POLL_SLEEP      = 1000 * time.Millisecond
	TEST_POLL_SLEEP = (POLL_SLEEP + (50 * time.Millisecond)) * 2
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
                                                              
        _//                                                   
     _//   _//                                                
    _//          _//    _// _//     _//    _/ _///_//   _//   
    _//        _//  _//  _//  _// _//  _//  _//    _// _//    
    _//       _//   _//  _//  _//_//   _//  _//      _///     
     _//   _//_//   _//  _//  _//_//   _//  _//       _//     
       _////    _// _///_///  _//  _// _///_///      _//      
                                                   _//        
                                                              `
)
