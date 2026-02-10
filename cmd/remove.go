package cmd

import (
	"fmt"

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

	// Filter out bare entries
	var filtered []git.WorktreeInfo
	for _, wt := range worktrees {
		if !wt.Bare {
			filtered = append(filtered, wt)
		}
	}

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

	if err := project.RunTeardownHooks(ctx, cfg, selected.Path, IsDryRun()); err != nil {
		ui.Warning("Teardown hooks failed: " + err.Error())
	}

	ui.Step("Removing worktree: " + selected.Branch)
	if err := runner.WorktreeRemove(ctx, selected.Path, force); err != nil {
		return err
	}

	if err := runner.BranchDelete(ctx, selected.Branch, false); err != nil {
		ui.Warning("Could not delete branch: " + err.Error())
	}

	ui.Success("Removed worktree: " + selected.Branch)
	return nil
}
