package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/project"
	"github.com/briankildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a worktree project in the current directory",
		Args:  cobra.NoArgs,
		RunE:  runInit,
	}
	cmd.Flags().Bool("force", false, "Overwrite existing configuration")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	dry := IsDryRun()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	if config.Exists(cwd) && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", config.ConfigFileName)
	}

	// Check for existing git state
	gitDir := filepath.Join(cwd, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		ui.Warning("Found existing .git directory. This project already has a standard git repository.")
	}

	bareDir := filepath.Join(cwd, config.DefaultGitDir)
	if info, err := os.Stat(bareDir); err == nil && info.IsDir() {
		ui.Info("Found existing .bare directory, will use as git dir.")
	}

	ui.Step("Creating project scaffold")
	if err := project.CreateScaffold(cwd, dry); err != nil {
		return err
	}

	ui.Step("Writing " + config.ConfigFileName)
	cfg := config.DefaultConfig()
	if dry {
		ui.DryRunNotice("write " + filepath.Join(cwd, config.ConfigFileName))
	} else {
		if err := cfg.Save(cwd); err != nil {
			return err
		}
	}

	ui.Success("Initialized worktree project")
	return nil
}
