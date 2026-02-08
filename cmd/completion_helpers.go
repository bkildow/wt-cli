package cmd

import (
	"context"
	"os"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/git"
	"github.com/briankildow/wt-cli/internal/project"
	"github.com/spf13/cobra"
)

func completeWorktreeNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := listWorktreeNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func listWorktreeNames() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projectRoot, err := project.FindRoot(cwd)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	runner := git.NewRunner(project.GitDirPath(projectRoot, cfg), false)
	worktrees, err := runner.WorktreeList(context.Background())
	if err != nil {
		return nil, err
	}

	var names []string
	for _, wt := range worktrees {
		if !wt.Bare {
			names = append(names, wt.Branch)
		}
	}

	return names, nil
}
