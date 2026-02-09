package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [branch]",
		Short: "Create a new worktree",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAdd,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, dry)

	var branch string
	if len(args) > 0 {
		branch = args[0]
	} else {
		prompter := &ui.InteractivePrompter{}
		branches, err := runner.ListRemoteBranches(ctx)
		if err != nil {
			return err
		}
		if len(branches) > 0 {
			branch, err = prompter.SelectBranch(branches)
			if err != nil {
				return err
			}
		} else {
			branch, err = prompter.InputString("Branch name", "feature/my-branch")
			if err != nil {
				return err
			}
		}
	}

	worktreePath := filepath.Join(projectRoot, "worktrees", branch)

	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("worktree already exists: worktrees/%s", branch)
	}

	hasRemote, err := runner.HasRemoteBranch(ctx, branch)
	if err != nil {
		return err
	}

	ui.Step("Adding worktree for branch: " + branch)
	if hasRemote {
		if err := runner.WorktreeAdd(ctx, worktreePath, branch); err != nil {
			return err
		}
	} else {
		if err := runner.WorktreeAddNew(ctx, worktreePath, branch); err != nil {
			return err
		}
	}

	vars := project.NewTemplateVars(worktreePath, branch)
	if err := project.Apply(projectRoot, worktreePath, dry, &vars); err != nil {
		return err
	}

	if err := project.RunSetupHooks(ctx, cfg, worktreePath, dry); err != nil {
		return err
	}

	ui.Success("Worktree created: worktrees/" + branch)
	return nil
}
