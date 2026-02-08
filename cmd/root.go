package cmd

import "github.com/spf13/cobra"

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
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.AddCommand(newCloneCmd())
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newRemoveCmd())
	rootCmd.AddCommand(newCdCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newInitCmd())
}

func Execute() error {
	return rootCmd.Execute()
}

func IsDryRun() bool {
	return dryRun
}
