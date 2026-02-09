package main

import (
	"os"

	"github.com/bkildow/wt-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
