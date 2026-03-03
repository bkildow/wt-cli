package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize wt in an existing git repository",
		Long:  "Wraps an existing git repository for worktree management.\nCreates .worktree.yml, shared/ directories, and the worktrees directory.",
		Args:  cobra.NoArgs,
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	dry := IsDryRun()

	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	// Check that .git is a directory (not a file, which would mean we're inside a worktree)
	gitPath := filepath.Join(projectRoot, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return fmt.Errorf("no .git directory found — use 'wt clone' for bare repo setup")
	}
	if !info.IsDir() {
		return fmt.Errorf(".git is a file, not a directory — this may already be a worktree of another repo")
	}

	// Refuse if already a wt project
	if config.Exists(projectRoot) {
		return fmt.Errorf("already a wt project (%s exists)", config.ConfigFileName)
	}

	cfg := config.DefaultConfig()
	cfg.GitDir = ".git"

	// Create scaffold directories
	ui.Step("Creating project scaffold")
	if err := project.CreateScaffold(projectRoot, &cfg, dry); err != nil {
		return err
	}

	// Write annotated config
	ui.Step("Writing " + config.ConfigFileName)
	if dry {
		ui.DryRunNotice("write " + filepath.Join(projectRoot, config.ConfigFileName))
	} else {
		if err := config.WriteAnnotatedWithValues(projectRoot, &cfg); err != nil {
			return err
		}
	}

	ui.Success("Initialized wt project in: " + projectRoot)
	ui.Info("  Your existing checkout is the main worktree.")
	ui.Info("  Use 'wt add <branch>' to create additional worktrees.")
	ui.Info("")
	ui.Info("  Consider adding to .gitignore:")
	ui.Info("    worktrees/")
	ui.Info("  Optionally ignore shared/ too (or commit it for team consistency).")

	return nil
}
