// Package main is the entry point for the wt CLI.
package main

import (
	"os"

	"github.com/bkildow/wt-cli/cmd"
	"github.com/bkildow/wt-cli/internal/ui"
)

func main() {
	if err := cmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
