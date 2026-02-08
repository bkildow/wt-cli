package project

import (
	"context"
	"testing"

	"github.com/briankildow/wt-cli/internal/config"
)

func TestRunPostCreateHooks(t *testing.T) {
	cfg := &config.Config{
		PostCreate: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunPostCreateHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunPostCreateHooks error: %v", err)
	}
}

func TestRunPostCreateHooksDryRun(t *testing.T) {
	cfg := &config.Config{
		PostCreate: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunPostCreateHooks(context.Background(), cfg, wt, true)
	if err != nil {
		t.Fatalf("RunPostCreateHooks dry-run error: %v", err)
	}
}

func TestRunPostCreateHooksFailure(t *testing.T) {
	cfg := &config.Config{
		PostCreate: []string{"false"},
	}
	wt := t.TempDir()

	err := RunPostCreateHooks(context.Background(), cfg, wt, false)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRunPostCreateHooksEmpty(t *testing.T) {
	cfg := &config.Config{}
	wt := t.TempDir()

	err := RunPostCreateHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunPostCreateHooks with empty hooks error: %v", err)
	}
}

func TestRunPostCreateHooksContinuesOnFailure(t *testing.T) {
	cfg := &config.Config{
		PostCreate: []string{"echo ok", "false", "echo still-runs"},
	}
	wt := t.TempDir()

	err := RunPostCreateHooks(context.Background(), cfg, wt, false)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}
