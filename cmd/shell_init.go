package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const bashFunction = `wt() {
  if [ "$1" = "cd" ] || [ "$1" = "add" ] || [ "$1" = "root" ]; then
    local dir
    dir="$(command wt "$@")"
    if [ -n "$dir" ]; then
      cd "$dir" || return
    fi
  else
    command wt "$@"
  fi
}
`

const zshFunction = `unalias wt 2>/dev/null
eval 'wt() {
  if [ "$1" = "cd" ] || [ "$1" = "add" ] || [ "$1" = "root" ]; then
    local dir
    dir="$(command wt "$@")"
    if [ -n "$dir" ]; then
      cd "$dir" || return
    fi
  else
    command wt "$@"
  fi
}'
`

const fishFunction = `function wt
  if test "$argv[1]" = "cd" -o "$argv[1]" = "add" -o "$argv[1]" = "root"
    set -l dir (command wt $argv)
    if test -n "$dir"
      cd "$dir"
    end
  else
    command wt $argv
  end
end
`

func newShellInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "shell-init [bash|zsh|fish]",
		Short:     "Print shell startup configuration (wrapper function + completions)",
		Long:      "Outputs shell code that sets up the wt cd wrapper function and tab completions. Add to your shell config with eval.",
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]

			// Print wrapper function
			switch shell {
			case "bash":
				fmt.Print(bashFunction)
			case "zsh":
				fmt.Print(zshFunction)
			case "fish":
				fmt.Print(fishFunction)
			}

			// Print completions
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			}

			return nil
		},
	}
}
