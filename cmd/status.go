package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/charmbracelet/lipgloss"
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

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	projectRoot, err := project.FindRoot(cwd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), IsDryRun())
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

	if len(filtered) == 0 {
		ui.Info("No worktrees found. Use 'wt add' to create one.")
		return nil
	}

	w := tabwriter.NewWriter(ui.Output, 0, 0, 2, ' ', 0)
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

		status := "clean"
		color := ui.ColorSuccess
		if dirty {
			status = "dirty"
			color = ui.ColorWarning
		}
		styledStatus := lipgloss.NewStyle().Foreground(color).Render(status)

		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", wt.Branch, relPath, shortHead, styledStatus, age)
	}
	return w.Flush()
}
