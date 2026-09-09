package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bkildow/wt-cli/internal/config"
)

func TestRunSetupHooks(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false, nil)
	if err != nil {
		t.Fatalf("RunSetupHooks error: %v", err)
	}
}

func TestRunSetupHooksDryRun(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), true, nil)
	if err != nil {
		t.Fatalf("RunSetupHooks dry-run error: %v", err)
	}
}

func TestRunSetupHooksFailure(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"false"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false, nil)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRunSetupHooksEmpty(t *testing.T) {
	cfg := &config.Config{}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false, nil)
	if err != nil {
		t.Fatalf("RunSetupHooks with empty hooks error: %v", err)
	}
}

func TestRunSetupHooksContinuesOnFailure(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo ok", "false", "echo still-runs"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false, nil)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRunTeardownHooks(t *testing.T) {
	cfg := &config.Config{
		Teardown: []string{"echo cleanup"},
	}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunTeardownHooks error: %v", err)
	}
}

func TestRunTeardownHooksEmpty(t *testing.T) {
	cfg := &config.Config{}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunTeardownHooks with empty hooks error: %v", err)
	}
}

func TestRunTeardownHooksFailure(t *testing.T) {
	cfg := &config.Config{
		Teardown: []string{"false"},
	}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err == nil {
		t.Fatal("expected error from failing teardown hook")
	}
}

func TestRunParallelSetupHooks(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{
		ParallelSetup: []string{
			"echo hello",
			"echo world",
		},
	}

	err := RunParallelSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunParallelSetupHooks error: %v", err)
	}
}

func TestRunParallelSetupHooksDryRun(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{
		ParallelSetup: []string{"echo hello", "echo world"},
	}

	err := RunParallelSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), true)
	if err != nil {
		t.Fatalf("RunParallelSetupHooks dry-run error: %v", err)
	}
}

func TestRunParallelSetupHooksEmpty(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{}

	err := RunParallelSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunParallelSetupHooks with empty hooks error: %v", err)
	}
}

func TestRunParallelSetupHooksFailure(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{
		ParallelSetup: []string{"echo ok", "false", "echo still-runs"},
	}

	err := RunParallelSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err == nil {
		t.Fatal("expected error from failing parallel setup hook")
	}
}

func TestRunParallelSetupHooksConcurrency(t *testing.T) {
	wt := t.TempDir()
	// Each command writes a file; verify all files exist afterward.
	cfg := &config.Config{
		ParallelSetup: []string{
			"touch " + filepath.Join(wt, "a.txt"),
			"touch " + filepath.Join(wt, "b.txt"),
			"touch " + filepath.Join(wt, "c.txt"),
		},
	}

	err := RunParallelSetupHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunParallelSetupHooks error: %v", err)
	}

	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if _, err := os.Stat(filepath.Join(wt, name)); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestRunParallelTeardownHooks(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{
		ParallelTeardown: []string{"echo cleanup1", "echo cleanup2"},
	}

	err := RunParallelTeardownHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err != nil {
		t.Fatalf("RunParallelTeardownHooks error: %v", err)
	}
}

func TestRunParallelTeardownHooksFailure(t *testing.T) {
	wt := t.TempDir()
	cfg := &config.Config{
		ParallelTeardown: []string{"echo ok", "false"},
	}

	err := RunParallelTeardownHooks(context.Background(), cfg, NewTemplateVars(wt, wt, "test"), false)
	if err == nil {
		t.Fatal("expected error from failing parallel teardown hook")
	}
}

func TestHooksReceiveWTEnv(t *testing.T) {
	root := t.TempDir()
	wt := filepath.Join(root, "worktrees", "feature", "Auth")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "env.txt")
	cfg := &config.Config{
		Setup:            []string{`echo "$WT_PROJECT_ROOT|$WT_WORKTREE_ID|$WT_WORKTREE_PATH|$WT_BRANCH_NAME|$(pwd)" > ` + out},
		ParallelTeardown: []string{`echo "$WT_WORKTREE_ID" >> ` + out},
		PostRemove:       []string{`echo "post:$(pwd)" >> ` + out},
	}
	vars := NewTemplateVars(root, wt, "feature/Auth")

	if err := RunSetupHooks(context.Background(), cfg, vars, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := RunParallelTeardownHooks(context.Background(), cfg, vars, false); err != nil {
		t.Fatal(err)
	}
	if err := RunPostRemoveHooks(context.Background(), cfg, vars, false); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	realWt, _ := filepath.EvalSymlinks(wt)
	realRoot, _ := filepath.EvalSymlinks(root)
	want := root + "|feature-auth|" + wt + "|feature/Auth|" + realWt + "\nfeature-auth\npost:" + realRoot + "\n"
	if got != want {
		t.Errorf("hook env mismatch\n got: %q\nwant: %q", got, want)
	}
}
