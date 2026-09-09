package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Captured from herdr 0.9.0 (see integrations/herdr/EVENT_PAYLOADS.log).
const herdrCreatedPayload = `{"event":"worktree_created","data":{"type":"worktree_created","workspace":{"workspace_id":"wX","number":3,"label":"herdr-hook-probe","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"wX:t1","agent_status":"unknown","worktree":{"repo_key":"/Users/me/Code/proj/.git","repo_name":"proj","repo_root":"/Users/me/Code/proj","checkout_path":"/Users/me/.herdr/worktrees/proj/feat","is_linked_worktree":true}},"worktree":{"path":"/Users/me/.herdr/worktrees/proj/feat","branch":"feat","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"open_workspace_id":"wX","label":"proj"}}}`

const herdrRemovedPayload = `{"event":"worktree_removed","data":{"type":"worktree_removed","workspace_id":"wX","workspace":{"workspace_id":"wX","worktree":{"repo_key":"/Users/me/Code/proj/.bare","repo_name":"proj","repo_root":"/Users/me/Code/proj/worktrees/main","checkout_path":"/Users/me/.herdr/worktrees/proj/feat","is_linked_worktree":true}},"worktree":{"path":"/Users/me/.herdr/worktrees/proj/feat","branch":"feat","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"label":"proj"},"forced":true}}`

func TestParseHerdrEvent_created(t *testing.T) {
	ev, err := parseHerdrEvent(herdrCreatedPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Event != herdrEventCreated {
		t.Errorf("event = %q, want %q", ev.Event, herdrEventCreated)
	}
	if ev.Data.Worktree.Path != "/Users/me/.herdr/worktrees/proj/feat" {
		t.Errorf("path = %q", ev.Data.Worktree.Path)
	}
	if ev.Data.Worktree.Branch != "feat" {
		t.Errorf("branch = %q", ev.Data.Worktree.Branch)
	}
	if ev.Data.Workspace.Worktree.RepoKey != "/Users/me/Code/proj/.git" {
		t.Errorf("repo_key = %q", ev.Data.Workspace.Worktree.RepoKey)
	}
	if ev.Data.Worktree.IsBare {
		t.Error("is_bare should be false")
	}
}

func TestParseHerdrEvent_removed(t *testing.T) {
	ev, err := parseHerdrEvent(herdrRemovedPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Event != herdrEventRemoved {
		t.Errorf("event = %q, want %q", ev.Event, herdrEventRemoved)
	}
	if !ev.Data.Forced {
		t.Error("forced should be true")
	}
	if ev.Data.Workspace.Worktree.RepoKey != "/Users/me/Code/proj/.bare" {
		t.Errorf("repo_key = %q", ev.Data.Workspace.Worktree.RepoKey)
	}
}

func TestParseHerdrEvent_invalid(t *testing.T) {
	if _, err := parseHerdrEvent("{nope"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func writeWorktreeYML(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".worktree.yml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveHerdrProjectRoot_bareLayout(t *testing.T) {
	root := t.TempDir()
	writeWorktreeYML(t, root)

	got, ok := resolveHerdrProjectRoot(filepath.Join(root, ".bare"), filepath.Join(root, "worktrees", "main"))
	if !ok || got != root {
		t.Errorf("got (%q, %v), want (%q, true)", got, ok, root)
	}
}

func TestResolveHerdrProjectRoot_dotGitLayout(t *testing.T) {
	root := t.TempDir()
	writeWorktreeYML(t, root)

	got, ok := resolveHerdrProjectRoot(filepath.Join(root, ".git"), root)
	if !ok || got != root {
		t.Errorf("got (%q, %v), want (%q, true)", got, ok, root)
	}
}

func TestResolveHerdrProjectRoot_fallsBackToRepoRoot(t *testing.T) {
	root := t.TempDir()
	writeWorktreeYML(t, root)
	nested := filepath.Join(root, "worktrees", "main")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	// repo_key points somewhere unexpected; repo_root is inside the project.
	got, ok := resolveHerdrProjectRoot(filepath.Join(t.TempDir(), "elsewhere.git"), nested)
	if !ok || got != root {
		t.Errorf("got (%q, %v), want (%q, true)", got, ok, root)
	}
}

func TestResolveHerdrProjectRoot_notWtProject(t *testing.T) {
	plain := t.TempDir()
	if _, ok := resolveHerdrProjectRoot(filepath.Join(plain, ".git"), plain); ok {
		t.Error("expected no project root for a repo without .worktree.yml")
	}
	if _, ok := resolveHerdrProjectRoot("", ""); ok {
		t.Error("expected no project root for empty inputs")
	}
}

func TestLoadHerdrHookContext_envMissing(t *testing.T) {
	t.Setenv(herdrEventEnv, "")
	if _, err := loadHerdrHookContext(herdrEventCreated); err == nil {
		t.Fatal("expected error when env is unset")
	}
}

func TestLoadHerdrHookContext_wrongEvent(t *testing.T) {
	t.Setenv(herdrEventEnv, herdrRemovedPayload)
	_, err := loadHerdrHookContext(herdrEventCreated)
	if err == nil || !strings.Contains(err.Error(), "unexpected herdr event") {
		t.Fatalf("expected wrong-event error, got %v", err)
	}
}

func TestLoadHerdrHookContext_notWtProjectIsQuiet(t *testing.T) {
	t.Setenv(herdrEventEnv, herdrCreatedPayload) // paths do not exist on this machine
	hctx, err := loadHerdrHookContext(herdrEventCreated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hctx != nil {
		t.Fatal("expected nil context for a non-wt repo")
	}
}

func TestHerdrManifest(t *testing.T) {
	m := herdrManifest("/opt/bin/wt")
	for _, want := range []string{
		`id = "wt"`,
		`on = "worktree.created"`,
		`on = "worktree.removed"`,
		`command = ["/opt/bin/wt", "herdr", "hook-worktree-created"]`,
		`command = ["/opt/bin/wt", "herdr", "hook-worktree-removed"]`,
		`min_herdr_version = "0.9.0"`,
	} {
		if !strings.Contains(m, want) {
			t.Errorf("manifest missing %q\n%s", want, m)
		}
	}
}

// newHerdrTestProject creates a wt project in the .git layout and returns its
// root. hooks is appended verbatim to .worktree.yml.
func newHerdrTestProject(t *testing.T, hooks string) string {
	t.Helper()
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	cfg := "version: 1\ngit_dir: .git\nworktree_dir: worktrees\nshared_dir: shared\n" + hooks
	if err := os.WriteFile(filepath.Join(root, ".worktree.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func herdrRemovedEventFor(root, worktreePath, branch string) string {
	return `{"event":"worktree_removed","data":{"type":"worktree_removed","workspace":{"worktree":{"repo_key":"` + root + `/.git","repo_root":"` + root + `"}},"worktree":{"path":"` + worktreePath + `","branch":"` + branch + `","is_bare":false},"forced":false}}`
}

func runRemovedHook(t *testing.T) error {
	t.Helper()
	cmd := newHerdrHookRemovedCmd()
	cmd.SetContext(context.Background())
	return runHerdrHookRemoved(cmd, nil)
}

func TestHerdrRemoved_skipsTeardownWhenDirGone_runsPostRemove(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "hooks.log")
	root := newHerdrTestProject(t,
		"teardown:\n  - 'echo teardown >> "+marker+"'\n"+
			"parallel_teardown:\n  - 'echo parallel >> "+marker+"'\n"+
			"post_remove:\n  - 'echo post_remove id=$WT_WORKTREE_ID pwd=$(pwd) >> "+marker+"'\n")
	gone := filepath.Join(t.TempDir(), "herdr", "proj", "feat-x") // never created

	t.Setenv(herdrStateDirEnv, t.TempDir())
	t.Setenv(herdrEventEnv, herdrRemovedEventFor(root, gone, "feat/x"))

	if err := runRemovedHook(t); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("post_remove hook did not run: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "teardown") || strings.Contains(got, "parallel") {
		t.Errorf("teardown hooks must not run when the worktree dir is gone:\n%s", got)
	}
	if !strings.Contains(got, "post_remove id=feat-x pwd="+root) {
		t.Errorf("post_remove should run from project root with WT_WORKTREE_ID set:\n%s", got)
	}
}

func TestHerdrRemoved_runsTeardownWhenDirPresent(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "hooks.log")
	root := newHerdrTestProject(t,
		"teardown:\n  - 'echo teardown pwd=$(pwd) >> "+marker+"'\n"+
			"post_remove:\n  - 'echo post_remove >> "+marker+"'\n")
	present := filepath.Join(t.TempDir(), "feat")
	if err := os.MkdirAll(present, 0o755); err != nil {
		t.Fatal(err)
	}
	present, _ = filepath.EvalSymlinks(present)

	t.Setenv(herdrStateDirEnv, t.TempDir())
	t.Setenv(herdrEventEnv, herdrRemovedEventFor(root, present, "feat"))

	if err := runRemovedHook(t); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(marker)
	got := string(data)
	if !strings.Contains(got, "teardown pwd="+present) {
		t.Errorf("teardown should run inside the still-present worktree:\n%s", got)
	}
	if strings.Contains(got, "post_remove") {
		t.Errorf("post_remove must not run when ordinary teardown ran:\n%s", got)
	}
}

func TestHerdrRemoved_stopsRunningSetup(t *testing.T) {
	root := newHerdrTestProject(t, "")
	gone := filepath.Join(t.TempDir(), "feat")
	t.Setenv(herdrStateDirEnv, t.TempDir())
	t.Setenv(herdrEventEnv, herdrRemovedEventFor(root, gone, "feat"))

	sleeper := exec.Command("sleep", "60")
	if err := sleeper.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sleeper.Process.Kill(); _ = sleeper.Wait() })

	if err := writeHerdrSetupPID(gone, sleeper.Process.Pid); err != nil {
		t.Fatal(err)
	}

	if err := runRemovedHook(t); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan struct{})
	go func() { _ = sleeper.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("removal hook did not stop the in-progress setup process")
	}
	if _, err := os.Stat(herdrSetupPIDPath(gone)); !os.IsNotExist(err) {
		t.Error("pid file should be removed after stopping setup")
	}
}

func TestHerdrSetupPIDPath_usesStateDirAndDistinguishesPaths(t *testing.T) {
	state := t.TempDir()
	t.Setenv(herdrStateDirEnv, state)
	a := herdrSetupPIDPath("/x/proj-a/feat")
	b := herdrSetupPIDPath("/x/proj-b/feat")
	if !strings.HasPrefix(a, filepath.Join(state, "setup")) {
		t.Errorf("pid path %q not under state dir", a)
	}
	if a == b {
		t.Error("same branch in different projects must not share a pid file")
	}
	if _, ok := readHerdrSetupPID("/nope"); ok {
		t.Error("missing pid file should read as absent")
	}
}
