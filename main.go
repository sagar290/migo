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

	fmt.Printf("config: %+v\n", config)
}
