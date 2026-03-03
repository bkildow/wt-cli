// Package project provides project-level operations including root detection and scaffold creation.
package project

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/ui"
)

func FindRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if config.Exists(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", config.ErrConfigNotFound
		}
		dir = parent
	}
}

func CreateScaffold(projectRoot string, cfg *config.Config, dryRun bool) error {
	dirs := []string{
		filepath.Join(projectRoot, "shared", "copy"),
		filepath.Join(projectRoot, "shared", "symlink"),
		WorktreesPath(projectRoot, cfg),
	}

	for _, dir := range dirs {
		if dryRun {
			ui.DryRunNotice("mkdir -p " + dir)
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}

var sshURLPattern = regexp.MustCompile(`[^/]+[:/]([^/]+/[^/]+?)(?:\.git)?$`)

func RepoNameFromURL(url string) string {
	// Handle SSH URLs: git@github.com:org/repo.git
	if matches := sshURLPattern.FindStringSubmatch(url); len(matches) > 1 {
		parts := strings.Split(matches[1], "/")
		return parts[len(parts)-1]
	}

	// Handle HTTPS URLs: https://github.com/org/repo.git
	base := filepath.Base(url)
	return strings.TrimSuffix(base, ".git")
}

func GitDirPath(projectRoot string, cfg *config.Config) string {
	return filepath.Join(projectRoot, cfg.GitDir)
}

func WorktreesPath(projectRoot string, cfg *config.Config) string {
	return filepath.Join(projectRoot, cfg.WorktreeDir)
}
