# herdr integration

[herdr](https://herdr.dev) owns worktree creation and removal (sidebar, `herdr worktree create/remove`).
Plugins cannot intercept those operations, but they can react to the
`worktree.created` and `worktree.removed` events. `wt herdr init` installs a
global herdr plugin whose two event hooks call back into `wt`:

| herdr event        | wt command                        | what it does                                                           |
|--------------------|-----------------------------------|------------------------------------------------------------------------|
| `worktree.created` | `wt herdr hook-worktree-created`  | `wt apply` for the new checkout, then `setup` / `parallel_setup` hooks |
| `worktree.removed` | `wt herdr hook-worktree-removed`  | stop in-flight setup, `post_remove` hooks, then `git worktree prune`   |

The manifest is written to `~/.config/wt/herdr-plugin/herdr-plugin.toml` and
linked with `herdr plugin link`. It is not per project: it fires for every repo
herdr touches and exits quietly when the repo has no `.worktree.yml`.

## Facts learned from herdr 0.9.0 (see `event-payloads.txt`)

- Event hooks are **not blocking**: `herdr worktree create` returns before the
  hook finishes, so setup may run in the hook process. Output is visible with
  `herdr plugin log list --plugin wt`.
- Hooks receive the payload in `HERDR_PLUGIN_EVENT_JSON` (not stdin). The
  fields wt uses are `data.worktree.path`, `data.worktree.branch`, and
  `data.workspace.worktree.repo_key` (the shared git dir, e.g. `<root>/.bare`).
- Checkouts default to `~/.herdr/worktrees/<repo>/<branch-slug>` (herdr
  `[worktrees] directory`), outside the project tree. The project root is
  resolved from `repo_key`, never by walking up from the checkout.
- `worktree.removed` fires **after** `git worktree remove`; the directory is
  already gone. Ordinary `teardown`/`parallel_teardown` hooks are skipped
  (they are written to run inside the checkout and would act on the main
  checkout in a `.git` layout). `post_remove` hooks run instead, from the
  project root, with `WT_PROJECT_ROOT`, `WT_WORKTREE_ID`, `WT_WORKTREE_PATH`,
  and `WT_BRANCH_NAME` exported.
- Event hooks are asynchronous, so a worktree can be removed while its setup
  is still running. The created hook records its PID under
  `$HERDR_PLUGIN_STATE_DIR/setup/`; the removed hook stops that process and
  waits for it before running `post_remove`.
- herdr does not delete the branch on removal; neither does wt here. Use
  `wt prune` for merged branches.
