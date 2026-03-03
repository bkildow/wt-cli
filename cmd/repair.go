package cmd

import (
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Rewrite worktree paths from absolute to relative",
		RunE:  runRepair,
	}
}

func runRepair(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	dry := IsDryRun()

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, dry)

	ui.Step("Repairing worktree paths to use relative references")
	if err := runner.WorktreeRepair(ctx); err != nil {
		return err
	}

	ui.Success("Worktree paths repaired")
	return nil
}
