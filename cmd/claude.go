package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bkildow/wt-cli/internal/claude"
	"github.com/bkildow/wt-cli/internal/config"
	"github.com/bkildow/wt-cli/internal/git"
	"github.com/bkildow/wt-cli/internal/project"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/spf13/cobra"
)

// hookPayload is the JSON structure Claude Code sends on stdin for hook events.
// See https://code.claude.com/docs/en/hooks for the payload schema.
type hookPayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	Name           string `json:"name"`          // WorktreeCreate: worktree slug
	WorktreePath   string `json:"worktree_path"` // WorktreeRemove: absolute path
}

func newClaudeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Claude Code integration",
		Long:  "Commands for integrating wt with Claude Code hooks.",
	}

	cmd.AddCommand(newClaudeInitCmd())
	cmd.AddCommand(newClaudeHookWorktreeCreateCmd())
	cmd.AddCommand(newClaudeHookWorktreeRemoveCmd())

	return cmd
}

func newClaudeInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Configure Claude Code hooks for this project",
		Long:  "Writes WorktreeCreate and WorktreeRemove hooks to .claude/settings.local.json so that Claude Code agents create worktrees through wt.",
		Args:  cobra.NoArgs,
		RunE:  runClaudeInit,
	}
	cmd.Flags().String("binary", "wt", "Path or name of the wt binary to use in hook commands")
	return cmd
}

func newClaudeHookWorktreeCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook-worktree-create",
		Short:  "Handle Claude Code WorktreeCreate hook",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE:   runClaudeHookWorktreeCreate,
	}
}

func newClaudeHookWorktreeRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook-worktree-remove",
		Short:  "Handle Claude Code WorktreeRemove hook",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE:   runClaudeHookWorktreeRemove,
	}
}

func runClaudeInit(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	wtBinary, _ := cmd.Flags().GetString("binary")

	// Write hooks to shared/symlink so all worktrees get a symlink via wt apply.
	sharedTarget := filepath.Join(projectRoot, cfg.SharedDir, "symlink")

	if claude.IsHooksConfigured(sharedTarget) {
		ui.Info("Claude Code hooks are already configured, updating...")
	}

	if err := claude.ConfigureHooks(sharedTarget, wtBinary); err != nil {
		return fmt.Errorf("failed to configure hooks: %w", err)
	}

	ui.Success("Configured Claude Code hooks in shared/symlink/.claude/settings.local.json")
	ui.Info("  WorktreeCreate -> " + wtBinary + " claude hook-worktree-create")
	ui.Info("  WorktreeRemove -> " + wtBinary + " claude hook-worktree-remove")

	// Apply to all existing worktrees so they get the symlink immediately.
	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, false)
	worktrees, err := runner.WorktreeList(ctx)
	if err != nil {
		ui.Warning("Could not list worktrees: " + err.Error())
		return nil
	}
	filtered := filterManagedWorktrees(worktrees, projectRoot)
	for _, wt := range filtered {
		vars := project.NewTemplateVars(projectRoot, wt.Path, wt.Branch)
		if _, err := project.Apply(projectRoot, wt.Path, cfg, false, &vars); err != nil {
			ui.Warning(fmt.Sprintf("Could not apply to worktree %s: %s", wt.Branch, err.Error()))
		}
	}
	ui.Success(fmt.Sprintf("Applied hooks to %d existing worktrees", len(filtered)))

	return nil
}

func runClaudeHookWorktreeCreate(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	hctx, err := loadHookContext()
	if err != nil {
		return err
	}

	projectRoot, cfg := hctx.projectRoot, hctx.cfg
	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, false)

	// Ensure git excludes are configured (non-fatal if sandbox blocks it).
	if err := project.EnsureGitExclude(gitDir, false); err != nil {
		ui.Warning("Could not configure git excludes: " + err.Error())
	}

	// Skip git fetch — Claude Code hooks run in a sandbox that restricts
	// writes to .bare/, and fetch requires network access. HasRemoteBranch
	// uses git branch -r (local only) which is sufficient.

	branch := hctx.payload.Name
	worktreePath := filepath.Join(project.WorktreesPath(projectRoot, cfg), branch)

	// If the worktree already exists, just return its path.
	if _, err := os.Stat(worktreePath); err == nil {
		fmt.Println(worktreePath)
		return nil
	}

	hasRemote, err := runner.HasRemoteBranch(ctx, branch)
	if err != nil {
		return fmt.Errorf("branch check failed: %w", err)
	}

	ui.Step("Adding worktree for branch: " + branch)
	if hasRemote {
		if err := runner.WorktreeAdd(ctx, worktreePath, branch); err != nil {
			return fmt.Errorf("worktree add failed: %w", err)
		}
	} else {
		if err := runner.WorktreeAddNew(ctx, worktreePath, branch); err != nil {
			return fmt.Errorf("worktree add (new branch) failed: %w", err)
		}
	}

	vars := project.NewTemplateVars(projectRoot, worktreePath, branch)
	result, err := project.Apply(projectRoot, worktreePath, cfg, false, &vars)
	if err != nil {
		return fmt.Errorf("apply shared files failed: %w", err)
	}

	msg := fmt.Sprintf("Worktree created: %s/%s (%d copied, %d symlinked)",
		cfg.WorktreeDir, branch, result.Copied, result.Symlinked)

	// Launch setup hooks in background if configured.
	// runSetupBackground prints the worktree path to stdout on its own.
	hasHooks := len(cfg.Setup) > 0 || len(cfg.ParallelSetup) > 0
	if hasHooks {
		if err := runSetupBackground(projectRoot, worktreePath, cfg, false, msg); err != nil {
			// Setup hook failure is non-fatal — the worktree is still usable.
			ui.Warning("Background setup failed to start: " + err.Error())
			fmt.Println(worktreePath)
		}
		return nil
	}

	// No hooks — print path directly.
	ui.Success(msg)
	fmt.Println(worktreePath)
	return nil
}

func runClaudeHookWorktreeRemove(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	hctx, err := loadHookContext()
	if err != nil {
		return err
	}

	projectRoot, cfg := hctx.projectRoot, hctx.cfg
	worktreePath := hctx.payload.WorktreePath

	// Derive branch name from the worktree path (everything after the worktrees dir).
	worktreesDir := project.WorktreesPath(projectRoot, cfg)
	branch, err := filepath.Rel(worktreesDir, worktreePath)
	if err != nil {
		return fmt.Errorf("cannot determine branch from worktree path: %w", err)
	}

	// Terminate any in-progress background setup.
	terminateBackgroundSetup(worktreePath, branch)

	// Run teardown hooks.
	if err := project.RunTeardownHooks(ctx, cfg, worktreePath, false); err != nil {
		ui.Warning("Teardown hooks failed: " + err.Error())
	}
	if err := project.RunParallelTeardownHooks(ctx, cfg, worktreePath, false); err != nil {
		ui.Warning("Parallel teardown hooks failed: " + err.Error())
	}

	gitDir := project.GitDirPath(projectRoot, cfg)
	runner := git.NewRunner(gitDir, false)

	// Force remove — Claude agents may have uncommitted changes.
	ui.Step("Removing worktree: " + branch)
	if err := runner.WorktreeRemove(ctx, worktreePath, true); err != nil {
		ui.Warning("Worktree remove failed: " + err.Error())
	}

	if err := runner.BranchDelete(ctx, branch, false); err != nil {
		ui.Warning("Could not delete branch: " + err.Error())
	}

	ui.Success("Removed worktree: " + branch)
	return nil
}

// hookContext bundles common state resolved during hook initialization.
type hookContext struct {
	payload     hookPayload
	projectRoot string
	cfg         *config.Config
}

// loadHookContext reads the JSON payload from stdin, validates it, resolves the
// project root, and loads the config. Both hook handlers share this setup.
func loadHookContext() (*hookContext, error) {
	payload, err := readHookPayload(os.Stdin)
	if err != nil {
		return nil, err
	}

	// Validate required fields based on event type.
	switch payload.HookEventName {
	case claude.HookWorktreeCreate:
		if payload.Name == "" {
			return nil, fmt.Errorf("name is required in WorktreeCreate payload")
		}
	case claude.HookWorktreeRemove:
		if payload.WorktreePath == "" {
			return nil, fmt.Errorf("worktree_path is required in WorktreeRemove payload")
		}
	}

	projectRoot, err := resolveProjectRoot(payload)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &hookContext{payload: payload, projectRoot: projectRoot, cfg: cfg}, nil
}

// readHookPayload parses the JSON hook payload from the given reader.
// Uses json.NewDecoder so it returns as soon as one JSON object is read,
// without waiting for EOF (which may never arrive from Claude Code's pipe).
func readHookPayload(r io.Reader) (hookPayload, error) {
	var payload hookPayload
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		if err == io.EOF {
			return hookPayload{}, fmt.Errorf("no payload received on stdin")
		}
		return hookPayload{}, fmt.Errorf("invalid JSON payload: %w", err)
	}
	return payload, nil
}

// resolveProjectRoot determines the project root from the hook payload.
func resolveProjectRoot(payload hookPayload) (string, error) {
	if payload.Cwd != "" {
		root, err := project.FindRoot(payload.Cwd)
		if err == nil {
			return root, nil
		}
	}

	return "", fmt.Errorf("could not find wt project root from payload (cwd=%q)", payload.Cwd)
}
