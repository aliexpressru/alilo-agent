package internal

import (
	"flag"
	"fmt"
	"os"

	"github.com/aliexpressru/alilo-agent/internal/tools/flag_tools"
	"github.com/aliexpressru/alilo-agent/pkg/utils"
)

func Start() {
	flag_tools.DeclarationCMDArg(&logFileName, &configFileName, &logLevel, &maxLogSize,
		&maxLogBackups, &maxLogAge, &serverPort)

	fmt.Printf("Launch arguments: len(%v); %v;\n", len(os.Args[1:]), os.Args[1:])
	flag.Parse()
	fmt.Printf("ConfigFileName: %v\n", configFileName)
	fmt.Printf("ServerPort: %v\n", serverPort)
	fmt.Printf("LogFileName: %v\n", logFileName)

	if !initLogger() {
		fmt.Println("Initialization logger error")
		return
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			fmt.Println("Error sync logger: ", err.Error())
		}
	}()
	if !initCfg() {
		fmt.Println("Initialization config error")
		return
	}

	fmt.Printf("IpAddress: %v\n", utils.FindOutTheIpAddress(logger))
	fmt.Printf("ServerPort: %v\n", cfg.ServerPort)

	if path, exists := os.LookupEnv("PATH"); exists {
		// Print the value of the environment variable
		fmt.Printf("PATH:\n'%v'\n", path)
	}
	handler, mux := addCors()
	PrepareServerAPI(mux)

	runEngineAgent(handler)
}
