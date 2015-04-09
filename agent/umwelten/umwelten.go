package umwelten

type Umwelten struct {
	Logo     string
	Env      string
	Prod     bool
	BaseUrl  string
	ConfPath string
}

const (
	defaultConfPath = "/etc/canary/canary.conf"
	testConfPath    = "agent/testdata/test2.conf"
	testLogo        = `		                                                      
        ********** ********  ******** **********      
       /////**/// /**/////  **////// /////**///       
           /**    /**      /**           /**          
           /**    /******* /*********    /**          
           /**    /**////  ////////**    /**          
           /**    /**             /**    /**          
           /**    /******** ********     /**          
           //     //////// ////////      //           
      ****     ****   *******   *******   ********    
     /**/**   **/**  **/////** /**////** /**/////     
     /**//** ** /** **     //**/**    /**/**          
     /** //***  /**/**      /**/**    /**/*******     
     /**  //*   /**/**      /**/**    /**/**////      
     /**   /    /**//**     ** /**    ** /**          
     /**        /** //*******  /*******  /********    
     //         //   ///////   ///////   ////////     
`
	strLogo = `
                                                              
        _//                                                   
     _//   _//                                                
    _//          _//    _// _//     _//    _/ _///_//   _//   
    _//        _//  _//  _//  _// _//  _//  _//    _// _//    
    _//       _//   _//  _//  _//_//   _//  _//      _///     
     _//   _//_//   _//  _//  _//_//   _//  _//       _//     
       _////    _// _///_///  _//  _// _///_///      _//      
                                                   _//        
                                                              `
	testURL = "http://localhost:8080"
)

var env = &Umwelten{}

func Init(env_str string) {
	env.Env = env_str
	if env_str != "test" && env_str != "debug" {
		env.Prod = true
	}

	if env.Prod {
		env.BaseUrl = "https://lolprod.example.com"
		env.Logo = strLogo
		env.ConfPath = defaultConfPath
	} else {
		env.BaseUrl = "http://localhost:8080"
		env.Logo = testLogo
		env.ConfPath = testConfPath
	}
}

func Fetch() *Umwelten {
	return env
}
