package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [branch]",
		Short: "Create a new worktree",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAdd,
	}
	cmd.Flags().Bool("skip-setup", false, "Skip running setup hooks after creating the worktree")
	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	dry := IsDryRun()

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, dry)

	ui.Step("Fetching all remotes")
	if err := runner.FetchAll(ctx); err != nil {
		return err
	}

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
				if ui.IsUserAbort(err) {
					return nil
				}
				return err
			}
		} else {
			branch, err = prompter.InputString("Branch name", "feature/my-branch")
			if err != nil {
				if ui.IsUserAbort(err) {
					return nil
				}
				return err
			}
		}
	}

	worktreePath := filepath.Join(project.WorktreesPath(projectRoot, cfg), branch)

	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("worktree already exists: %s/%s", cfg.WorktreeDir, branch)
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

	vars := project.NewTemplateVars(projectRoot, worktreePath, branch)
	result, err := project.Apply(projectRoot, worktreePath, dry, &vars)
	if err != nil {
		return err
	}

	var setupErr error
	skipSetup, _ := cmd.Flags().GetBool("skip-setup")
	if !skipSetup {
		setupErr = project.RunSetupHooks(ctx, cfg, worktreePath, dry)
		if pErr := project.RunParallelSetupHooks(ctx, cfg, worktreePath, dry); pErr != nil {
			setupErr = errors.Join(setupErr, pErr)
		}
	}

	msg := fmt.Sprintf("Worktree created: %s/%s (%d copied, %d symlinked)",
		cfg.WorktreeDir, branch, result.Copied, result.Symlinked)
	if setupErr != nil {
		ui.Warning(msg + " — setup hooks failed")
	} else {
		ui.Success(msg)
	}
	fmt.Println(worktreePath)
	return nil
}
