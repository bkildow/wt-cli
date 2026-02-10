package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
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
			if ui.IsUserAbort(err) {
				return nil
			}
			return err
		}
		for _, wt := range filtered {
			if wt.Branch == name {
				selected = wt
				break
			}
		}
	}

	// Determine editor: config > $EDITOR > auto-detect
	var editorBinary string

	// 1. Config file
	if cfg.Editor != "" {
		if _, err := exec.LookPath(cfg.Editor); err != nil {
			return fmt.Errorf("configured editor not found: %s", cfg.Editor)
		}
		editorBinary = cfg.Editor
	}

	// 2. $EDITOR environment variable
	if editorBinary == "" {
		if env := os.Getenv("EDITOR"); env != "" {
			if _, err := exec.LookPath(env); err != nil {
				return fmt.Errorf("$EDITOR not found: %s", env)
			}
			editorBinary = env
		}
	}

	// 3. Auto-detect known editors
	if editorBinary == "" {
		var available []string
		for _, e := range knownEditors {
			if _, err := exec.LookPath(e.Binary); err == nil {
				available = append(available, e.Binary)
			}
		}
		switch len(available) {
		case 0:
			return fmt.Errorf("no editor found: set 'editor' in .worktree.yml or $EDITOR")
		case 1:
			editorBinary = available[0]
		default:
			prompter := &ui.InteractivePrompter{}
			editorBinary, err = prompter.SelectEditor(available)
			if err != nil {
				if ui.IsUserAbort(err) {
					return nil
				}
				return err
			}
		}
	}

	if IsDryRun() {
		ui.DryRunNotice(fmt.Sprintf("%s %s", editorBinary, selected.Path))
		return nil
	}

	ui.Step(fmt.Sprintf("Opening %s in %s", selected.Branch, editorBinary))
	c := exec.Command(editorBinary, selected.Path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
