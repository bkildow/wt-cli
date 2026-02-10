package cmd

import (
	"errors"
	"os"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
)

// findProjectRoot resolves the current directory and walks up to find the project root.
// Prints a friendly error if no project is found.
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := project.FindRoot(cwd)
	if errors.Is(err, config.ErrConfigNotFound) {
		ui.Error("Not a wt project (no .worktree.yml found)")
		ui.Info("  Run 'wt clone <repo-url>' to create one.")
	}
	return root, err
}

// loadProject finds the project root and loads its config.
// Prints a friendly error if no project is found.
func loadProject() (string, *config.Config, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return "", nil, err
	}

	return root, cfg, nil
}
