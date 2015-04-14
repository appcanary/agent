package umwelten

// env vars
const (
	PROD_URL = "https://lolprod.example.com"
	DEV_URL  = "http://localhost:9999"

	DEFAULT_CONF_PATH = "/etc/canary/canary.conf"
	DEV_CONF_PATH     = "../test/data/test2.conf"
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
