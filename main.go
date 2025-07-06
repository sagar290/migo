package main

import (
	"fmt"
	"github.com/sagar290/migo/migo"
)

func main() {

	config, err := migo.LoadConfig()
	if err != nil {
		return
	}

	migoInstance, err := migo.NewMigo(config, migo.NewTracker(config))
	if err != nil {
		return
	}

	fmt.Printf("config: %+v\n", config)
	fmt.Printf("config: %+v\n", migoInstance)
}
