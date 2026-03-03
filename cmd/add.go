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

	skipSetup, _ := cmd.Flags().GetBool("skip-setup")
	if !skipSetup {
		if err := project.RunSetupHooks(ctx, cfg, worktreePath, dry); err != nil {
			return err
		}
	}

	ui.Success("Worktree created: worktrees/" + branch)
	fmt.Println(worktreePath)
	return nil
}
