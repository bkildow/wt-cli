package cmd

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var version = "dev"

var dryRun bool

var rootCmd = &cobra.Command{
	Use:           "wt",
	Short:         "A smarter git worktree workflow",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Detect color capabilities against stderr so colors work even when
	// stdout is piped (e.g. wt cd under the shell wrapper function).
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.AddCommand(newAgentsCmd())
	rootCmd.AddCommand(newCloneCmd())
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
}

func Execute() error {
	return rootCmd.Execute()
}

func IsDryRun() bool {
	return dryRun
}
