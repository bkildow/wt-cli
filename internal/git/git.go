package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/briankildow/wt-cli/internal/ui"
)

type WorktreeInfo struct {
	Path   string
	Branch string
	Head   string
	Bare   bool
}

type Git interface {
	Run(ctx context.Context, args ...string) (string, error)
	CloneBare(ctx context.Context, url, dest string) error
	ConfigureRemoteFetch(ctx context.Context) error
	Fetch(ctx context.Context, remote string) error
	ListRemoteBranches(ctx context.Context) ([]string, error)
	WorktreeAdd(ctx context.Context, path, branch string) error
	WorktreeList(ctx context.Context) ([]WorktreeInfo, error)
}

type Runner struct {
	GitDir string
	DryRun bool
}

func NewRunner(gitDir string, dryRun bool) *Runner {
	return &Runner{GitDir: gitDir, DryRun: dryRun}
}

func (r *Runner) Run(ctx context.Context, args ...string) (string, error) {
	fullArgs := append([]string{"--git-dir", r.GitDir}, args...)
	cmdStr := "git " + strings.Join(fullArgs, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return "", nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r *Runner) CloneBare(ctx context.Context, url, dest string) error {
	args := []string{"clone", "--bare", url, dest}
	cmdStr := "git " + strings.Join(args, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone --bare: %w\n%s", err, stderr.String())
	}

	return nil
}

func (r *Runner) ConfigureRemoteFetch(ctx context.Context) error {
	_, err := r.Run(ctx, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	return err
}

func (r *Runner) Fetch(ctx context.Context, remote string) error {
	_, err := r.Run(ctx, "fetch", remote)
	return err
}

func (r *Runner) ListRemoteBranches(ctx context.Context) ([]string, error) {
	output, err := r.Run(ctx, "branch", "-r")
	if err != nil {
		return nil, err
	}

	return parseRemoteBranches(output), nil
}

func parseRemoteBranches(output string) []string {
	if output == "" {
		return nil
	}

	var branches []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip HEAD pointer lines like "origin/HEAD -> origin/main"
		if strings.Contains(line, "->") {
			continue
		}
		// Strip "origin/" prefix
		branch := strings.TrimPrefix(line, "origin/")
		branches = append(branches, branch)
	}

	return branches
}

func (r *Runner) WorktreeAdd(ctx context.Context, path, branch string) error {
	_, err := r.Run(ctx, "worktree", "add", path, branch)
	return err
}

func (r *Runner) WorktreeList(ctx context.Context) ([]WorktreeInfo, error) {
	output, err := r.Run(ctx, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseWorktreeList(output), nil
}

func parseWorktreeList(output string) []WorktreeInfo {
	if output == "" {
		return nil
	}

	var worktrees []WorktreeInfo
	var current WorktreeInfo

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.Bare = true
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
		}
	}

	// Append the last entry if output doesn't end with a blank line
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}
