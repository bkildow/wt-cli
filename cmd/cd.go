package cmd

import (
	"fmt"
	"os"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func newCdCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "cd [name]",
		Short:             "Print worktree path for shell navigation",
		Long:              "Prints the absolute path of a worktree. Use with: cd \"$(wt cd)\"",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE:              runCd,
	}
}

func runCd(cmd *cobra.Command, args []string) error {
	// Detect if stdout is a terminal (wrapper pipes stdout, so TTY means no wrapper)
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		ui.Info("Tip: wt cd prints a path but can't change your directory directly.")
		ui.Info("  Run: eval \"$(wt shell-init zsh)\"  (or bash|fish) to set up the wrapper.")
	}

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), IsDryRun())
	worktrees, err := runner.WorktreeList(cmd.Context())
	if err != nil {
		return err
	}

	// Filter out bare entries and build lookup
	var filtered []git.WorktreeInfo
	var names []string
	for _, wt := range worktrees {
		if !wt.Bare {
			filtered = append(filtered, wt)
			names = append(names, wt.Branch)
		}
	}

	if len(filtered) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	var selected string
	if len(args) > 0 {
		selected = args[0]
	} else {
		prompter := &ui.InteractivePrompter{}
		selected, err = prompter.SelectWorktree(names)
		if err != nil {
			if ui.IsUserAbort(err) {
				return nil
			}
			return err
		}
	}

	for _, wt := range filtered {
		if wt.Branch == selected {
			fmt.Println(wt.Path)
			return nil
		}
	}

	return fmt.Errorf("worktree not found: %s", selected)
}
