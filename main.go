package main

import (
	"fmt"
	"github.com/sagar290/migo/cmd"
	"os"
)

func main() {

	cmd.Init()

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

}
