package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"

	"github.com/spf13/cobra"
)

func newRunSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "_run-setup",
		Hidden: true,
		RunE:   runRunSetup,
	}
	cmd.Flags().String("worktree-path", "", "Path to the worktree")
	cmd.Flags().String("project-root", "", "Path to the project root")
	return cmd
}

func runRunSetup(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	worktreePath, _ := cmd.Flags().GetString("worktree-path")
	projectRoot, _ := cmd.Flags().GetString("project-root")

	if worktreePath == "" || projectRoot == "" {
		return fmt.Errorf("--worktree-path and --project-root are required")
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	logFile, err := os.Create(project.SetupLogPath(worktreePath))
	if err != nil {
		return err
	}
	defer func() { _ = logFile.Close() }()

	// Redirect all UI output to the log file.
	ui.Output = logFile

	hooksTotal := len(cfg.Setup) + len(cfg.ParallelSetup)

	state := &project.SetupState{
		Status:     project.SetupRunning,
		PID:        os.Getpid(),
		StartedAt:  time.Now(),
		HooksTotal: hooksTotal,
		LogFile:    project.SetupLogPath(worktreePath),
	}
	if err := project.WriteSetupState(worktreePath, state); err != nil {
		return err
	}

	var setupErr error

	// Run serial hooks with progress tracking.
	onProgress := func(index int, cmdStr string, hookErr error) {
		state.HooksCompleted = index + 1
		_ = project.WriteSetupState(worktreePath, state)
	}
	setupErr = project.RunSetupHooks(ctx, cfg, worktreePath, false, onProgress)

	// Run parallel hooks as a batch.
	if pErr := project.RunParallelSetupHooks(ctx, cfg, worktreePath, false); pErr != nil {
		setupErr = errors.Join(setupErr, pErr)
	}
	state.HooksCompleted = hooksTotal

	if setupErr != nil {
		state.Status = project.SetupFailed
		state.Error = setupErr.Error()
	} else {
		state.Status = project.SetupComplete
	}
	state.CompletedAt = time.Now()

	elapsed := ui.FormatDuration(state.CompletedAt.Sub(state.StartedAt))
	if setupErr != nil {
		ui.Warning("Worktree setup failed after " + elapsed)
	} else {
		ui.Success("Worktree setup completed in " + elapsed)
	}

	return project.WriteSetupState(worktreePath, state)
}
