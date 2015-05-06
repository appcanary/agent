package umwelten

import "time"

// consts can't be outputs of functions?
var DEV_CONF_PATH string
var DEV_CONF_FILE string

var DEV_VAR_PATH string
var DEV_VAR_FILE string

// env vars
const (
	PROD_URL = "https://lolprod.example.com"
	DEV_URL  = "http://localhost:9999"

	DEFAULT_CONF_PATH = "/etc/canary/"
	DEFAULT_CONF_FILE = DEFAULT_CONF_PATH + "canary.conf"
	DEFAULT_VAR_PATH  = "/var/db/canary/"
	DEFAULT_VAR_FILE  = DEFAULT_VAR_PATH + "server.conf"

	DEFAULT_HEARTBEAT_DURATION = 1 * time.Hour
	DEV_HEARTBEAT_DURATION     = 10 * time.Second
)

// api endpoints
const (
	API_VERSION   = "/v1/agent/"
	API_HEARTBEAT = API_VERSION + "heartbeat"
	API_SERVERS   = API_VERSION + "servers"
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
