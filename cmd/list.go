package cmd

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), IsDryRun())
	worktrees, err := runner.WorktreeList(cmd.Context())
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
		fmt.Fprintf(w, "  %s\t%s\t%s\n", wt.Branch, relPath, shortHead)
	}
	return w.Flush()
}
