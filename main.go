package main

import (
	"fmt"
	"os"

	"github.com/stateio/canary-agent/agent"
)

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
)

func main() {
	done := make(chan os.Signal, 1)

	// what config are we loading?
	filename := ""
	env := os.Getenv("CANARY_ENV")
	if env == "test" || env == "debug" {
		filename = testConfPath
	} else {
		filename = defaultConfPath
	}

	fmt.Println(strLogo)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println("We need to implement getting the conf info from the user")
		return
	}

	conf := agent.NewConfFromFile(filename)
	a := agent.NewAgent(conf)
	defer a.CloseWatches()

	//	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
