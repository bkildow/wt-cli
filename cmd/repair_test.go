package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasBareFalse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"simple false", "[core]\n\tbare = false\n", true},
		{"simple true", "[core]\n\tbare = true\n", false},
		{"missing key", "[core]\n\tfoo = bar\n", false},
		{"key in other section", "[remote]\n\tbare = false\n", false},
		{"case-insensitive section", "[CORE]\n\tBARE = False\n", true},
		{"comments and blanks", "# comment\n\n[core]\n; another\nbare = false\n", true},
		{"section switch back", "[core]\nbare = true\n[remote]\nbare = false\n", false},
		{"section switch and override", "[remote]\nbare = false\n[core]\nbare = false\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasBareFalse(tt.in)
			if got != tt.want {
				t.Errorf("hasBareFalse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestWorktreeConfigPath_LinkedWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Build a real bare repo + linked worktree so we can resolve a real .git
	// file pointer.
	srcDir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", srcDir)
	run("-C", srcDir, "config", "user.email", "t@t")
	run("-C", srcDir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(srcDir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("-C", srcDir, "add", ".")
	run("-C", srcDir, "commit", "-m", "init")

	projectDir := t.TempDir()
	bareDir := filepath.Join(projectDir, ".bare")
	run("clone", "--bare", srcDir, bareDir)
	run("--git-dir", bareDir, "config", "extensions.worktreeConfig", "true")

	// Discover the default branch — git's clone --bare leaves HEAD pointing at it.
	headOut, err := exec.Command("git", "--git-dir", bareDir, "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("symbolic-ref HEAD: %v", err)
	}
	branch := strings.TrimSpace(string(headOut))

	wtPath := filepath.Join(projectDir, "worktrees", branch)
	run("--git-dir", bareDir, "worktree", "add", "--relative-paths", wtPath, branch)

	got, err := worktreeConfigPath(wtPath)
	if err != nil {
		t.Fatalf("worktreeConfigPath: %v", err)
	}
	want := filepath.Join(bareDir, "worktrees", branch, "config.worktree")
	if resolveForTest(got) != resolveForTest(want) {
		t.Errorf("worktreeConfigPath = %q, want %q", got, want)
	}
}

func resolveForTest(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return filepath.Clean(p)
}
