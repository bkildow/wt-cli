# wt

A CLI for managing git worktree-based development workflows. Clone once as a bare repo, then spin up isolated worktrees per branch with shared config files, symlinks, and template variables.

## Features

- **Bare-repo workflow** — no `.git` at project root; all worktrees live under `worktrees/`
- **Shared files** — copy per-worktree configs or symlink heavy directories (node_modules, vendor) once
- **Template variables** — `${WORKTREE_NAME}`, `${DATABASE_NAME}`, etc. substituted in copied files
- **Interactive by default** — branch/worktree pickers when arguments are omitted
- **Setup/teardown hooks** — run commands automatically when creating or removing worktrees
- **Editor integration** — open worktrees in your preferred editor ($EDITOR, config, or auto-detect)
- **Shell completions** — tab-complete worktree names in bash, zsh, and fish
- **Dry-run support** — preview every destructive operation with `--dry-run`

## Install

```bash
go install github.com/bkildow/wt-cli/cmd/wt@latest
```

Or build from source:

```bash
git clone https://github.com/bkildow/wt-cli.git
cd wt-cli
go build -o wt ./cmd/wt
# move wt to somewhere in your $PATH
```

## Quick Start

```bash
# Clone a repo into a bare worktree project
wt clone git@github.com:org/repo.git
cd repo

# Create a worktree for a feature branch
wt add feature/auth

# Navigate to it (see Shell Integration below for cd support)
cd "$(wt cd feature/auth)"

# See all worktrees
wt list

# When done, clean up merged branches
wt prune
```

## Commands

| Command | Description |
|---------|-------------|
| `wt clone <url> [name]` | Clone a repo as a bare worktree project |
| `wt add [branch]` | Create a new worktree for a branch |
| `wt list` | List all worktrees |
| `wt remove [name]` | Remove a worktree and its branch |
| `wt cd [name]` | Print worktree path for shell navigation |
| `wt apply [name]` | Apply shared files to a worktree |
| `wt open [name]` | Open a worktree in an IDE |
| `wt status` | Show status of all worktrees |
| `wt sync` | Fetch and pull all worktrees |
| `wt prune` | Remove worktrees with fully merged branches |
| `wt config init` | Generate annotated `.worktree.yml` with documentation |
| `wt agents` | Print AI agent workflow instructions |
| `wt completion <shell>` | Generate shell completion script |

### wt clone

```bash
wt clone <url> [name]        # Clone repo as bare worktree project
wt clone <url> --dry-run     # Preview without executing
```

Clones as a bare repo and writes `.worktree.yml`. Optionally prompts to create an initial worktree.

### wt config init

```bash
wt config init               # Generate annotated .worktree.yml (backs up existing)
wt config init --update      # Merge existing values into annotated template
```

Generates a `.worktree.yml` with documentation comments for every field. If a config already exists, it is backed up to `.worktree.yml.bak` first. Use `--update` to preserve your existing values while adding documentation comments.

### wt add

```bash
wt add feature/auth          # Create worktree for branch
wt add                       # Interactive branch picker
```

Detects whether the branch exists remotely or creates a new local branch. Applies shared files and runs setup hooks.

### wt remove

```bash
wt remove feature/auth       # Remove worktree and branch
wt remove --force            # Skip uncommitted changes check
```

Runs teardown hooks before removing the worktree directory.

### wt cd

```bash
cd "$(wt cd feature/auth)"   # Navigate to worktree
wt cd                        # Interactive picker
```

Prints the absolute path to stdout. When run without a shell wrapper, `wt cd` prints a hint about setting one up. See [Shell Integration](#shell-integration) for details.

### wt apply

```bash
wt apply feature/auth        # Apply shared files to one worktree
wt apply --all               # Apply to all worktrees
```

Copies files from `shared/copy/` (with template substitution) and creates symlinks from `shared/symlink/`.

### wt open

```bash
wt open feature/auth         # Open in editor
wt open                      # Interactive picker
```

Editor resolution order: `editor` field in `.worktree.yml` > `$EDITOR` env var > auto-detect (Cursor, VS Code, Zed).

### wt status

```bash
wt status
```

Shows branch, path, commit hash, dirty/clean status, and last commit age for all worktrees.

### wt sync

```bash
wt sync                      # Fetch + pull all clean worktrees
wt sync --rebase             # Use rebase instead of merge
```

Skips dirty worktrees. Shows summary of updated/skipped/failed counts.

### wt prune

```bash
wt prune                     # Remove worktrees with merged branches
wt prune --force             # Skip confirmation
```

Compares branches against the default branch (main/master).

### wt agents

```bash
wt agents                    # Print AI workflow guide to stdout
wt agents > AGENTS.md        # Save as a file in your project
```

Outputs structured workflow instructions for AI coding assistants to understand how to use `wt` in non-interactive mode.

### wt completion

```bash
wt completion bash > /etc/bash_completion.d/wt
wt completion zsh > "${fpath[1]}/_wt"
wt completion fish > ~/.config/fish/completions/wt.fish
```

## How It Works

`wt` organizes a project like this:

```
project/
├── .bare/                   # Bare git repository (no working tree)
├── .worktree.yml            # Project configuration
├── shared/
│   ├── copy/                # Files copied into each worktree
│   │   └── .env.example     # Supports ${TEMPLATE_VARS}
│   └── symlink/             # Shared resources symlinked from worktrees
│       ├── node_modules/
│       └── vendor/
└── worktrees/
    ├── main/                # Each branch gets its own directory
    ├── feature-auth/
    └── feature-ui/
```

**Why a bare repo?** Standard `git worktree` puts the primary checkout at the repo root, mixing repo files with worktree management. A bare repo at `.bare/` keeps the root clean — it only holds configuration and shared resources.

**Copy vs Symlink:** Files in `shared/copy/` are duplicated into each worktree (useful for `.env` files that vary per branch). Files in `shared/symlink/` are symlinked (useful for large directories like `node_modules` you only want to install once).

## Configuration

### .worktree.yml

```yaml
version: 1
git_dir: .bare
editor: cursor
setup:
  - "npm install"
  - "cp .env.example .env"
teardown:
  - "docker compose down"
```

| Field | Description | Default |
|-------|-------------|---------|
| `version` | Config version | `1` |
| `git_dir` | Path to bare repository | `.bare` |
| `editor` | Preferred editor binary name | (auto-detect) |
| `setup` | Commands to run after creating a worktree | `[]` |
| `teardown` | Commands to run before removing a worktree | `[]` |

### Template Variables

Files in `shared/copy/` support these variables:

| Variable | Example (branch: `feature/auth`) |
|----------|----------------------------------|
| `${WORKTREE_NAME}` | `feature-auth` |
| `${WORKTREE_PATH}` | `/path/to/worktrees/feature-auth` |
| `${BRANCH_NAME}` | `feature/auth` |
| `${DATABASE_NAME}` | `feature_auth` |

Binary files are detected and skipped automatically.

## Shell Integration

### Navigation Wrapper

`wt cd` prints a path to stdout. To make `wt cd` actually change your directory, add a shell wrapper:

**Bash / Zsh** (add to `~/.bashrc` or `~/.zshrc`):

```bash
wt() {
  if [ "$1" = "cd" ]; then
    shift
    local dir
    dir="$(command wt cd "$@")"
    if [ -n "$dir" ]; then
      cd "$dir" || return
    fi
  else
    command wt "$@"
  fi
}
```

**Fish** (add to `~/.config/fish/config.fish`):

```fish
function wt
  if test "$argv[1]" = "cd"
    set -l dir (command wt cd $argv[2..])
    if test -n "$dir"
      cd "$dir"
    end
  else
    command wt $argv
  end
end
```

Or generate it with `wt cd --shell-function bash|zsh|fish`.

### Tab Completions

```bash
# Bash
wt completion bash > /etc/bash_completion.d/wt

# Zsh
wt completion zsh > "${fpath[1]}/_wt"

# Fish
wt completion fish > ~/.config/fish/completions/wt.fish
```

Worktree names are completed for `add`, `remove`, `cd`, `apply`, and `open`.

## License

MIT
