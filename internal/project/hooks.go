package project

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/ui"
)

// RunSetupHooks executes each command in cfg.Setup inside the
// worktree directory. Failures are logged but do not stop subsequent hooks.
func RunSetupHooks(ctx context.Context, cfg *config.Config, worktreePath string, dryRun bool) error {
	if len(cfg.Setup) == 0 {
		return nil
	}

	failCount := 0
	for _, cmdStr := range cfg.Setup {
		ui.Step("Running: " + cmdStr)

		if dryRun {
			ui.DryRunNotice("exec: " + cmdStr)
			continue
		}

		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = ui.Output
		cmd.Stderr = ui.Output

		if err := cmd.Run(); err != nil {
			ui.Error("Failed: " + cmdStr + ": " + err.Error())
			failCount++
			continue
		}
		ui.Success("Completed: " + cmdStr)
	}

	if failCount > 0 {
		return fmt.Errorf("%d setup hook(s) failed", failCount)
	}
	return nil
}

// RunTeardownHooks executes each command in cfg.Teardown inside the
// worktree directory. Failures are logged but do not stop subsequent hooks.
func RunTeardownHooks(ctx context.Context, cfg *config.Config, worktreePath string, dryRun bool) error {
	if len(cfg.Teardown) == 0 {
		return nil
	}

	failCount := 0
	for _, cmdStr := range cfg.Teardown {
		ui.Step("Running: " + cmdStr)

		if dryRun {
			ui.DryRunNotice("exec: " + cmdStr)
			continue
		}

		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = ui.Output
		cmd.Stderr = ui.Output

		if err := cmd.Run(); err != nil {
			ui.Error("Failed: " + cmdStr + ": " + err.Error())
			failCount++
			continue
		}
		ui.Success("Completed: " + cmdStr)
	}

	if failCount > 0 {
		return fmt.Errorf("%d teardown hook(s) failed", failCount)
	}
	return nil
}
