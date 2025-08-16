package main

import (
	"fmt"
	"github.com/sagar290/migo/cmd"
	"github.com/sagar290/migo/src"
	"log"
	"os"
)

func main() {

	config, err := src.LoadConfig()
	if err != nil {
		log.Println("⚠️ config error:", err)
		return
	}

	migoInstance, err := src.NewMigo(config, src.NewTracker(config))
	if err != nil {
		return
	}

	cmd.Init(migoInstance)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	//fmt.Printf("config: %+v\n", config)
	//fmt.Printf("config: %+v\n", migoInstance)

}
