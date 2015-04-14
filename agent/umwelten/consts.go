package umwelten

import "path/filepath"

// consts can't be outputs of functions?
var DEV_CONF_PATH, _ = filepath.Abs("test/data/test2.conf")

// env vars
const (
	PROD_URL = "https://lolprod.example.com"
	DEV_URL  = "http://localhost:9999"

	DEFAULT_CONF_PATH = "/etc/canary/canary.conf"
)

// api endpoints
const (
	API_VERSION   = "/v1/agent/"
	API_HEARTBEAT = API_VERSION + "heartbeat/"
	API_SERVERS   = API_VERSION + "servers/"
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
