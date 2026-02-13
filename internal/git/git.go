package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bkildow/wt-cli/internal/ui"
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
	HasRemoteBranch(ctx context.Context, branch string) (bool, error)
	WorktreeAdd(ctx context.Context, path, branch string) error
	WorktreeAddNew(ctx context.Context, path, branch string) error
	WorktreeRemove(ctx context.Context, path string, force bool) error
	WorktreeList(ctx context.Context) ([]WorktreeInfo, error)
	WorktreePrune(ctx context.Context) error
	BranchDelete(ctx context.Context, branch string, force bool) error
	IsWorktreeDirty(ctx context.Context, worktreePath string) (bool, error)
	IsBranchMerged(ctx context.Context, branch, target string) (bool, error)
	FetchAll(ctx context.Context) error
	GetDefaultBranch(ctx context.Context) (string, error)
	GetLastCommitAge(ctx context.Context, worktreePath string) (string, error)
	GetBehindCount(ctx context.Context, worktreePath string) (int, error)
	Pull(ctx context.Context, worktreePath string) error
	PullRebase(ctx context.Context, worktreePath string) error
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

func (r *Runner) HasRemoteBranch(ctx context.Context, branch string) (bool, error) {
	branches, err := r.ListRemoteBranches(ctx)
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		if b == branch {
			return true, nil
		}
	}
	return false, nil
}

func (r *Runner) WorktreeAdd(ctx context.Context, path, branch string) error {
	_, err := r.Run(ctx, "worktree", "add", path, branch)
	return err
}

func (r *Runner) WorktreeAddNew(ctx context.Context, path, branch string) error {
	_, err := r.Run(ctx, "worktree", "add", "-b", branch, path, "HEAD")
	return err
}

func (r *Runner) WorktreeRemove(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := r.Run(ctx, args...)
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

func (r *Runner) BranchDelete(ctx context.Context, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.Run(ctx, "branch", flag, branch)
	return err
}

func (r *Runner) IsWorktreeDirty(ctx context.Context, worktreePath string) (bool, error) {
	args := []string{"-C", worktreePath, "status", "--porcelain"}
	cmdStr := "git " + strings.Join(args, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return false, nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return parseDirtyStatus(stdout.String()), nil
}

func parseDirtyStatus(output string) bool {
	return strings.TrimSpace(output) != ""
}

func (r *Runner) IsBranchMerged(ctx context.Context, branch, target string) (bool, error) {
	args := []string{"--git-dir", r.GitDir, "merge-base", "--is-ancestor", branch, target}
	cmdStr := "git " + strings.Join(args, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return true, nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return true, nil
}

func (r *Runner) FetchAll(ctx context.Context) error {
	_, err := r.Run(ctx, "fetch", "--all")
	return err
}

func (r *Runner) GetDefaultBranch(ctx context.Context) (string, error) {
	output, err := r.Run(ctx, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		return parseDefaultBranch(output), nil
	}

	if _, err := r.Run(ctx, "show-ref", "--verify", "--quiet", "refs/heads/main"); err == nil {
		return "main", nil
	}

	if _, err := r.Run(ctx, "show-ref", "--verify", "--quiet", "refs/heads/master"); err == nil {
		return "master", nil
	}

	return "", fmt.Errorf("could not determine default branch")
}

func (r *Runner) WorktreePrune(ctx context.Context) error {
	_, err := r.Run(ctx, "worktree", "prune")
	return err
}

func (r *Runner) GetLastCommitAge(ctx context.Context, worktreePath string) (string, error) {
	args := []string{"-C", worktreePath, "log", "-1", "--format=%cr"}
	cmdStr := "git " + strings.Join(args, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return "unknown", nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r *Runner) GetBehindCount(ctx context.Context, worktreePath string) (int, error) {
	args := []string{"-C", worktreePath, "rev-list", "--count", "HEAD..@{upstream}"}
	cmdStr := "git " + strings.Join(args, " ")

	if r.DryRun {
		ui.DryRunNotice(cmdStr)
		return 0, nil
	}

	ui.Command(cmdStr)
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "no upstream") || strings.Contains(stderrStr, "unknown revision") {
			return 0, nil
		}
		return 0, fmt.Errorf("%s: %w\n%s", cmdStr, err, stderrStr)
	}

	return parseBehindCount(stdout.String()), nil
}

func (r *Runner) Pull(ctx context.Context, worktreePath string) error {
	args := []string{"-C", worktreePath, "pull"}
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
		return fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return nil
}

func (r *Runner) PullRebase(ctx context.Context, worktreePath string) error {
	args := []string{"-C", worktreePath, "pull", "--rebase"}
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
		return fmt.Errorf("%s: %w\n%s", cmdStr, err, stderr.String())
	}

	return nil
}

func parseDefaultBranch(output string) string {
	s := strings.TrimSpace(output)
	return strings.TrimPrefix(s, "refs/remotes/origin/")
}

func parseBranchList(output string) []string {
	if output == "" {
		return nil
	}

	var branches []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "* ")
		branches = append(branches, line)
	}

	return branches
}

func parseBehindCount(output string) int {
	n, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0
	}
	return n
}
