package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const agentsMarkdown = `# AGENTS.md — AI Workflow Guide for wt

## Overview

wt is a CLI for git worktree-based development. It wraps a bare git repository
and creates isolated worktrees under a worktrees/ directory with shared config
files, symlinks, and template variable substitution.

## Important: Non-Interactive Usage

wt is interactive by default — commands launch pickers when arguments are
omitted. AI agents MUST always pass explicit arguments to avoid interactive
prompts.

    # Wrong (launches interactive picker):
    wt add
    wt remove
    wt cd

    # Correct (explicit arguments):
    wt add feature/auth
    wt remove feature/auth --force
    wt cd feature/auth

## Project Structure

    project/
      .bare/              # Bare git repository (no .git at root)
      .worktree.yml       # Project configuration
      shared/
        copy/             # Files copied into each new worktree
        symlink/          # Directories symlinked into each new worktree
      worktrees/
        main/             # Each branch gets its own directory
        feature-auth/
        feature-ui/

## Command Reference

### Clone a project

    wt clone <url> [name]
    wt clone <url> --dry-run          # Preview without executing

### Create a worktree

    wt add <branch>                   # Detects remote or creates new branch

### List worktrees

    wt list

### Remove a worktree

    wt remove <name> --force          # Use --force to skip confirmation

### Get worktree path

    wt cd <name>                      # Prints path to stdout (does NOT cd)

### Navigate to a worktree

    cd "$(wt cd <name>)"              # Use shell substitution to cd

### Apply shared files

    wt apply <name>                   # Apply to one worktree
    wt apply --all                    # Apply to all worktrees

### Open in editor

    wt open <name>

### Show status of all worktrees

    wt status

### Fetch and pull all worktrees

    wt sync                           # Pull all clean worktrees
    wt sync --rebase                  # Use rebase instead of merge

### Remove worktrees with merged branches

    wt prune --force                  # Use --force to skip confirmation

### Preview any command safely

    wt --dry-run <command> [args]

### Configuration management

    wt config init                    # Generate annotated .worktree.yml
    wt config init --update           # Preserve existing values

## Common Workflows

### Starting a new project

    wt clone git@github.com:org/repo.git
    cd repo
    wt add feature/my-feature
    cd "$(wt cd feature/my-feature)"

### Creating a feature branch

    wt add feature/my-feature
    cd "$(wt cd feature/my-feature)"

### Checking project state

    wt status
    wt list

### Cleaning up after merge

    wt sync
    wt prune --force

### Applying shared file changes

    wt apply --all

## Configuration (.worktree.yml)

    version: 1
    git_dir: .bare
    editor: cursor
    setup:
      - "npm install"
    teardown:
      - "docker compose down"

Fields:
- version: Config version (always 1)
- git_dir: Path to bare repository (default: .bare)
- editor: Preferred editor binary name (default: auto-detect)
- setup: Commands run after creating a worktree
- teardown: Commands run before removing a worktree

## Template Variables

Files in shared/copy/ support these substitution variables:

- ${WORKTREE_NAME} — branch with / replaced by - (e.g. feature-auth)
- ${WORKTREE_PATH} — absolute path to the worktree
- ${BRANCH_NAME} — original branch name (e.g. feature/auth)
- ${DATABASE_NAME} — lowercase with - and . replaced by _ (e.g. feature_auth)

Binary files are detected and skipped automatically.

## Key Caveats

1. wt cd prints a path — it does not change directory. Always use:
   cd "$(wt cd <name>)"
2. There is no .git at the project root. The bare repo lives at .bare/.
3. Use --force with wt remove and wt prune to skip interactive confirmation.
4. Use --dry-run to safely preview any destructive operation.
5. The project root is identified by .worktree.yml — look for this file.
6. Run git commands inside the worktree directory, not the project root.
7. Worktree directories live at worktrees/<branch-name>/ under the project root.
`

func newAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "agents",
		Short: "Print AI agent workflow instructions",
		Long:  "Outputs workflow instructions for AI tools to understand how to use wt effectively.\nPipe to a file to create an AGENTS.md: wt agents > AGENTS.md",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(agentsMarkdown)
			return nil
		},
	}
}
