package project

import (
	"context"
	"testing"

	"github.com/briankildow/wt-cli/internal/config"
)

func TestRunSetupHooks(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunSetupHooks error: %v", err)
	}
}

func TestRunSetupHooksDryRun(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo hello"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, wt, true)
	if err != nil {
		t.Fatalf("RunSetupHooks dry-run error: %v", err)
	}
}

func TestRunSetupHooksFailure(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"false"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, wt, false)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRunSetupHooksEmpty(t *testing.T) {
	cfg := &config.Config{}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunSetupHooks with empty hooks error: %v", err)
	}
}

func TestRunSetupHooksContinuesOnFailure(t *testing.T) {
	cfg := &config.Config{
		Setup: []string{"echo ok", "false", "echo still-runs"},
	}
	wt := t.TempDir()

	err := RunSetupHooks(context.Background(), cfg, wt, false)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRunTeardownHooks(t *testing.T) {
	cfg := &config.Config{
		Teardown: []string{"echo cleanup"},
	}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunTeardownHooks error: %v", err)
	}
}

func TestRunTeardownHooksEmpty(t *testing.T) {
	cfg := &config.Config{}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, wt, false)
	if err != nil {
		t.Fatalf("RunTeardownHooks with empty hooks error: %v", err)
	}
}

func TestRunTeardownHooksFailure(t *testing.T) {
	cfg := &config.Config{
		Teardown: []string{"false"},
	}
	wt := t.TempDir()

	err := RunTeardownHooks(context.Background(), cfg, wt, false)
	if err == nil {
		t.Fatal("expected error from failing teardown hook")
	}
}
