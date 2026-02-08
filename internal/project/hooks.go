package project

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/briankildow/wt-cli/internal/config"
	"github.com/briankildow/wt-cli/internal/ui"
)

// RunPostCreateHooks executes each command in cfg.PostCreate inside the
// worktree directory. Failures are logged but do not stop subsequent hooks.
func RunPostCreateHooks(ctx context.Context, cfg *config.Config, worktreePath string, dryRun bool) error {
	if len(cfg.PostCreate) == 0 {
		return nil
	}

	failCount := 0
	for _, cmdStr := range cfg.PostCreate {
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
		return fmt.Errorf("%d hook(s) failed", failCount)
	}
	return nil
}
