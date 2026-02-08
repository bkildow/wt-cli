package cmd

import (
	"fmt"
	"os"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/git"
	"github.com/briankildow/wt-cli/internal/project"
	"github.com/briankildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

const bashZshFunction = `wt() {
  if [ "$1" = "cd" ]; then
    shift
    local dir
    dir="$(command wt cd "$@")"
    if [ -n "$dir" ]; then
      cd "$dir" || return
    fi
  else
    command wt "$@"
  fi
}
`

const fishFunction = `function wt
  if test "$argv[1]" = "cd"
    set -l dir (command wt cd $argv[2..])
    if test -n "$dir"
      cd "$dir"
    end
  else
    command wt $argv
  end
end
`

func newCdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cd [name]",
		Short:             "Print worktree path for shell navigation",
		Long:              "Prints the absolute path of a worktree. Use with: cd \"$(wt cd)\"",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE:              runCd,
	}
	cmd.Flags().String("shell-function", "", "Print shell wrapper function (bash|zsh|fish)")
	return cmd
}

func printShellFunction(shell string) error {
	switch shell {
	case "bash", "zsh":
		fmt.Print(bashZshFunction)
	case "fish":
		fmt.Print(fishFunction)
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}
	return nil
}

func runCd(cmd *cobra.Command, args []string) error {
	shell, _ := cmd.Flags().GetString("shell-function")
	if shell != "" {
		return printShellFunction(shell)
	}

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
