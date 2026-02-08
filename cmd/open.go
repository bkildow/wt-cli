package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/git"
	"github.com/briankildow/wt-cli/internal/project"
	"github.com/briankildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

var knownEditors = []struct{ Name, Binary string }{
	{"Cursor", "cursor"},
	{"VS Code", "code"},
	{"Zed", "zed"},
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "open [name]",
		Short:             "Open a worktree in an IDE",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE:              runOpen,
	}
	cmd.Flags().Bool("cursor", false, "Open in Cursor")
	cmd.Flags().Bool("code", false, "Open in VS Code")
	cmd.Flags().Bool("zed", false, "Open in Zed")
	cmd.MarkFlagsMutuallyExclusive("cursor", "code", "zed")
	return cmd
}

func runOpen(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("no worktrees found")
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

	// Determine editor
	var editorBinary string

	// Check flags first
	for _, e := range knownEditors {
		flagSet, _ := cmd.Flags().GetBool(e.Binary)
		if flagSet {
			editorBinary = e.Binary
			break
		}
	}

	// Auto-detect if no flag set
	if editorBinary == "" {
		var available []string
		for _, e := range knownEditors {
			if _, err := exec.LookPath(e.Binary); err == nil {
				available = append(available, e.Binary)
			}
		}

		switch len(available) {
		case 0:
			return fmt.Errorf("no supported editor found (install cursor, code, or zed)")
		case 1:
			editorBinary = available[0]
		default:
			prompter := &ui.InteractivePrompter{}
			editorBinary, err = prompter.SelectEditor(available)
			if err != nil {
				return err
			}
		}
	}

	if IsDryRun() {
		ui.DryRunNotice(fmt.Sprintf("%s %s", editorBinary, selected.Path))
		return nil
	}

	ui.Step(fmt.Sprintf("Opening %s in %s", selected.Branch, editorBinary))
	return exec.Command(editorBinary, selected.Path).Start()
}
