// Package cmd implements the CLI commands for the wt worktree manager.
package cmd

import (
	"fmt"
	"os"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/bkildow/wt-cli/internal/ui"
	"github.com/charmbracelet/colorprofile"
	lipglossv1 "github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var dryRun bool

var rootCmd = &cobra.Command{
	Use:           "wt",
	Short:         "A smarter git worktree workflow",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if theme := os.Getenv("WT_THEME"); theme != "" {
			ui.ApplyTheme(theme)
		}
		return nil
	},
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)

	// Detect color capabilities against stderr so colors work even when
	// stdout is piped (e.g. wt cd under the shell wrapper function).
	//
	// lipgloss v1: huh v0.8.0 uses v1 styles internally. Without this,
	// v1's default renderer probes stdout (no-TTY when piped) and strips
	// all picker colors.
	lipglossv1.SetDefaultRenderer(lipglossv1.NewRenderer(os.Stderr))
	// lipgloss v2: override the default Writer (targets stdout) so any
	// code using lipgloss.Println / lipgloss.Sprint detects against stderr.
	lipgloss.Writer = colorprofile.NewWriter(os.Stderr, os.Environ())

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVar(&ui.Verbose, "verbose", false, "Show git commands being executed")
	rootCmd.AddCommand(newAgentsCmd())
	rootCmd.AddCommand(newCloneCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newRemoveCmd())
	rootCmd.AddCommand(newCdCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newShellInitCmd())
	rootCmd.AddCommand(newOpenCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newPruneCmd())
	rootCmd.AddCommand(newRootCmd())
}

func Execute() error {
	return rootCmd.Execute()
}

func IsDryRun() bool {
	return dryRun
}
