package cmd

import (
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of all worktrees",
		Args:  cobra.NoArgs,
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	if len(filtered) == 0 {
		ui.Info("No worktrees found. Use 'wt add' to create one.")
		return nil
	}

	ui.Heading("Worktree Status")

	t := ui.NewTable().Headers("BRANCH", "PATH", "COMMIT", "STATUS", "AGE")
	for _, wt := range filtered {
		relPath, err := filepath.Rel(projectRoot, wt.Path)
		if err != nil {
			relPath = wt.Path
		}

		shortHead := wt.Head
		if len(shortHead) > 7 {
			shortHead = shortHead[:7]
		}

		dirty, err := runner.IsWorktreeDirty(ctx, wt.Path)
		if err != nil {
			return err
		}

		age, err := runner.GetLastCommitAge(ctx, wt.Path)
		if err != nil {
			return err
		}

		styledStatus := ui.StyleSuccess.Render("clean")
		if dirty {
			styledStatus = ui.StyleWarning.Render("dirty")
		}

		t.Row(wt.Branch, relPath, shortHead, styledStatus, age)
	}
	ui.PrintTable(t)
	return nil
}
