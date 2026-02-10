package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove worktrees with fully merged branches",
		Args:  cobra.NoArgs,
		RunE:  runPrune,
	}
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return cmd
}

func runPrune(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	cwd, _ := os.Getwd()
	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), IsDryRun())

	defaultBranch, err := runner.GetDefaultBranch(ctx)
	if err != nil {
		return err
	}

	worktrees, err := runner.WorktreeList(ctx)
	if err != nil {
		return err
	}

	var filtered []git.WorktreeInfo
	for _, wt := range worktrees {
		if !wt.Bare {
			filtered = append(filtered, wt)
		}
	}

	// Resolve current worktree path for comparison
	currentPath, _ := filepath.EvalSymlinks(cwd)

	var pruneable []git.WorktreeInfo
	for _, wt := range filtered {
		if wt.Branch == defaultBranch {
			continue
		}

		wtPath, _ := filepath.EvalSymlinks(wt.Path)
		if wtPath == currentPath {
			continue
		}

		merged, err := runner.IsBranchMerged(ctx, wt.Branch, defaultBranch)
		if err != nil {
			ui.Warning(fmt.Sprintf("%s: could not check merge status: %s", wt.Branch, err))
			continue
		}

		if merged {
			pruneable = append(pruneable, wt)
		}
	}

	if len(pruneable) == 0 {
		ui.Info("No merged worktrees to prune.")
		return nil
	}

	ui.Step("Merged worktrees:")
	for _, wt := range pruneable {
		relPath, err := filepath.Rel(projectRoot, wt.Path)
		if err != nil {
			relPath = wt.Path
		}
		fmt.Fprintf(ui.Output, "  %s  %s\n", wt.Branch, relPath)
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !IsDryRun() {
		prompter := &ui.InteractivePrompter{}
		confirmed, err := prompter.Confirm(fmt.Sprintf("Remove %d merged worktree(s)?", len(pruneable)))
		if err != nil {
			if ui.IsUserAbort(err) {
				return nil
			}
			return err
		}
		if !confirmed {
			ui.Info("Cancelled.")
			return nil
		}
	}

	var removed int
	for _, wt := range pruneable {
		ui.Step("Removing worktree: " + wt.Branch)
		if err := runner.WorktreeRemove(ctx, wt.Path, false); err != nil {
			ui.Warning(fmt.Sprintf("Could not remove worktree %s: %s", wt.Branch, err))
			continue
		}

		if err := runner.BranchDelete(ctx, wt.Branch, false); err != nil {
			ui.Warning("Could not delete branch: " + err.Error())
		}

		removed++
	}

	if err := runner.WorktreePrune(ctx); err != nil {
		ui.Warning("Could not prune worktree metadata: " + err.Error())
	}

	ui.Success(fmt.Sprintf("Pruned %d worktree(s)", removed))
	return nil
}
