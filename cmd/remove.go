package cmd

import (
	"fmt"
	"syscall"
	"time"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove [name]",
		Short:             "Remove a worktree and its branch",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE:              runRemove,
	}
	cmd.Flags().Bool("force", false, "Remove even if worktree has uncommitted changes")
	cmd.Flags().Bool("skip-teardown", false, "Skip running teardown hooks before removing the worktree")
	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), IsDryRun())

	worktrees, err := runner.WorktreeList(ctx)
	if err != nil {
		return err
	}

	filtered := filterManagedWorktrees(worktrees, projectRoot)
	mainBranch := cfg.MainBranchOrDefault()

	if len(filtered) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	var selected git.WorktreeInfo
	if len(args) > 0 {
		found := false
		for _, wt := range filtered {
			if wt.Branch == args[0] {
				selected = wt
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("worktree not found: %s", args[0])
		}
	} else {
		names := make([]string, len(filtered))
		for i, wt := range filtered {
			names[i] = wt.Branch
		}
		prompter := &ui.InteractivePrompter{}
		name, err := prompter.SelectWorktree(names)
		if err != nil {
			if ui.IsUserAbort(err) {
				return nil
			}
			return err
		}
		for _, wt := range filtered {
			if wt.Branch == name {
				selected = wt
				break
			}
		}
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		dirty, err := runner.IsWorktreeDirty(ctx, selected.Path)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("worktree has uncommitted changes (use --force to override)")
		}
	}

	// Terminate any in-progress background setup before teardown.
	terminateBackgroundSetup(selected.Path, selected.Branch)

	skipTeardown, _ := cmd.Flags().GetBool("skip-teardown")
	if !skipTeardown {
		if err := project.RunTeardownHooks(ctx, cfg, selected.Path, IsDryRun()); err != nil {
			ui.Warning("Teardown hooks failed: " + err.Error())
		}
		if err := project.RunParallelTeardownHooks(ctx, cfg, selected.Path, IsDryRun()); err != nil {
			ui.Warning("Parallel teardown hooks failed: " + err.Error())
		}
	}

	isMainBranch := selected.Branch == mainBranch

	if isMainBranch {
		ui.Step("Removing worktree: " + selected.Branch + " (branch preserved in bare repo)")
	} else {
		ui.Step("Removing worktree: " + selected.Branch)
	}

	if err := runner.WorktreeRemove(ctx, selected.Path, force); err != nil {
		return err
	}

	if !isMainBranch {
		if err := runner.BranchDelete(ctx, selected.Branch, false); err != nil {
			ui.Warning("Could not delete branch: " + err.Error())
		}
	}

	ui.Success("Removed worktree: " + selected.Branch)
	return nil
}

// terminateBackgroundSetup kills an in-progress background setup process.
func terminateBackgroundSetup(worktreePath, branch string) {
	state, err := project.ReadSetupState(worktreePath)
	if err != nil || state == nil {
		return
	}
	if state.Status != project.SetupRunning || !project.IsProcessAlive(state.PID) {
		return
	}

	ui.Warning(fmt.Sprintf("Terminating in-progress setup for %s (PID %d)", branch, state.PID))
	_ = syscall.Kill(state.PID, syscall.SIGTERM)

	// Poll for exit, then force-kill if the process doesn't terminate.
	deadline := time.After(2 * time.Second)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if !project.IsProcessAlive(state.PID) {
				return
			}
		case <-deadline:
			if project.IsProcessAlive(state.PID) {
				_ = syscall.Kill(state.PID, syscall.SIGKILL)
			}
			return
		}
	}
}
