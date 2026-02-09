package cmd

import (
	"fmt"
	"os"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "apply [name]",
		Short:             "Apply shared files to a worktree",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE:              runApply,
	}
	cmd.Flags().Bool("all", false, "Apply to all worktrees")
	return cmd
}

func runApply(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	dry := IsDryRun()

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

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), dry)

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

	all, _ := cmd.Flags().GetBool("all")

	if all {
		for _, wt := range filtered {
			ui.Step("Applying to: " + wt.Branch)
			vars := project.NewTemplateVars(wt.Path, wt.Branch)
			if err := project.Apply(projectRoot, wt.Path, dry, &vars); err != nil {
				return err
			}
		}
		ui.Success(fmt.Sprintf("Applied shared files to %d worktree(s)", len(filtered)))
		return nil
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
			return err
		}
		for _, wt := range filtered {
			if wt.Branch == name {
				selected = wt
				break
			}
		}
	}

	vars := project.NewTemplateVars(selected.Path, selected.Branch)
	if err := project.Apply(projectRoot, selected.Path, dry, &vars); err != nil {
		return err
	}

	ui.Success("Applied shared files to: " + selected.Branch)
	return nil
}
