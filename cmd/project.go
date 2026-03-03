package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
)

// findProjectRoot resolves the current directory and walks up to find the project root.
// Prints a friendly error if no project is found.
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := project.FindRoot(cwd)
	if errors.Is(err, config.ErrConfigNotFound) {
		ui.Error("Not a wt project (no .worktree.yml found)")
		ui.Info("  Run 'wt clone <repo-url>' to create one, or 'wt init' inside an existing repo.")
	}
	return root, err
}

// loadProject finds the project root and loads its config.
// Prints a friendly error if no project is found.
func loadProject() (string, *config.Config, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return "", nil, err
	}

	return root, cfg, nil
}

// filterManagedWorktrees returns only worktrees that wt manages — those
// created under the worktrees directory. This excludes bare entries (from
// wt clone setups) and the main working tree at the project root (from
// wt init setups).
func filterManagedWorktrees(worktrees []git.WorktreeInfo, projectRoot string) []git.WorktreeInfo {
	absRoot := resolvePathBest(projectRoot)
	var filtered []git.WorktreeInfo
	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		// Fast path: exact string match avoids syscall
		if wt.Path == projectRoot || resolvePathBest(wt.Path) == absRoot {
			continue
		}
		filtered = append(filtered, wt)
	}
	return filtered
}

// resolvePathBest tries to resolve symlinks; falls back to filepath.Clean.
func resolvePathBest(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return filepath.Clean(p)
}
