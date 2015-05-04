package umwelten

import (
	"path/filepath"
	"time"
)

// consts can't be outputs of functions?
var DEV_CONF_PATH, _ = filepath.Abs("test/data/")
var DEV_CONF_FILE = filepath.Join(DEV_CONF_PATH, "test2.conf")

var DEV_VAR_PATH, _ = filepath.Abs("test/var")
var DEV_VAR_FILE = filepath.Join(DEV_VAR_PATH, "server.conf")

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
