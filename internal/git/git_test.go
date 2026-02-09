package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bkildow/wt-cli/internal/ui"
)

func TestParseRemoteBranches(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "typical output",
			input:  "  origin/main\n  origin/develop\n  origin/feature/login",
			expect: []string{"main", "develop", "feature/login"},
		},
		{
			name:   "with HEAD pointer",
			input:  "  origin/HEAD -> origin/main\n  origin/main\n  origin/develop",
			expect: []string{"main", "develop"},
		},
		{
			name:   "empty output",
			input:  "",
			expect: nil,
		},
		{
			name:   "single branch",
			input:  "  origin/main",
			expect: []string{"main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRemoteBranches(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("got %v, want %v", got, tt.expect)
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("branch[%d] = %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestParseWorktreeList(t *testing.T) {
	input := `worktree /home/user/project/.bare
HEAD abc1234567890
branch refs/heads/main
bare

worktree /home/user/project/worktrees/develop
HEAD def4567890123
branch refs/heads/develop

`

	got := parseWorktreeList(input)
	if len(got) != 2 {
		t.Fatalf("got %d worktrees, want 2", len(got))
	}

	if got[0].Path != "/home/user/project/.bare" {
		t.Errorf("worktree[0].Path = %q", got[0].Path)
	}
	if got[0].Branch != "main" {
		t.Errorf("worktree[0].Branch = %q, want %q", got[0].Branch, "main")
	}
	if !got[0].Bare {
		t.Error("worktree[0].Bare should be true")
	}

	if got[1].Path != "/home/user/project/worktrees/develop" {
		t.Errorf("worktree[1].Path = %q", got[1].Path)
	}
	if got[1].Branch != "develop" {
		t.Errorf("worktree[1].Branch = %q, want %q", got[1].Branch, "develop")
	}
	if got[1].Bare {
		t.Error("worktree[1].Bare should be false")
	}
}

func TestDryRunMode(t *testing.T) {
	// Redirect UI output to discard
	ui.Output = os.Stderr

	runner := NewRunner("/nonexistent", true)
	ctx := context.Background()

	// These should not error because dry-run skips execution
	out, err := runner.Run(ctx, "status")
	if err != nil {
		t.Errorf("dry-run Run returned error: %v", err)
	}
	if out != "" {
		t.Errorf("dry-run Run returned output: %q", out)
	}

	if err := runner.CloneBare(ctx, "https://example.com/repo.git", "/tmp/dest"); err != nil {
		t.Errorf("dry-run CloneBare returned error: %v", err)
	}

	if err := runner.ConfigureRemoteFetch(ctx); err != nil {
		t.Errorf("dry-run ConfigureRemoteFetch returned error: %v", err)
	}

	if err := runner.Fetch(ctx, "origin"); err != nil {
		t.Errorf("dry-run Fetch returned error: %v", err)
	}

	// HasRemoteBranch (dry-run returns false since ListRemoteBranches returns empty)
	hasBranch, err := runner.HasRemoteBranch(ctx, "main")
	if err != nil {
		t.Errorf("dry-run HasRemoteBranch returned error: %v", err)
	}
	if hasBranch {
		t.Errorf("dry-run HasRemoteBranch should return false")
	}

	// WorktreeAddNew
	if err := runner.WorktreeAddNew(ctx, "/tmp/wt", "feature-x"); err != nil {
		t.Errorf("dry-run WorktreeAddNew returned error: %v", err)
	}

	// WorktreeRemove
	if err := runner.WorktreeRemove(ctx, "/tmp/wt", false); err != nil {
		t.Errorf("dry-run WorktreeRemove returned error: %v", err)
	}
	if err := runner.WorktreeRemove(ctx, "/tmp/wt", true); err != nil {
		t.Errorf("dry-run WorktreeRemove (force) returned error: %v", err)
	}

	// BranchDelete
	if err := runner.BranchDelete(ctx, "old-branch", false); err != nil {
		t.Errorf("dry-run BranchDelete returned error: %v", err)
	}
	if err := runner.BranchDelete(ctx, "old-branch", true); err != nil {
		t.Errorf("dry-run BranchDelete (force) returned error: %v", err)
	}

	// IsWorktreeDirty
	dirty, err := runner.IsWorktreeDirty(ctx, "/tmp/wt")
	if err != nil {
		t.Errorf("dry-run IsWorktreeDirty returned error: %v", err)
	}
	if dirty {
		t.Errorf("dry-run IsWorktreeDirty should return false")
	}

	// IsBranchMerged
	merged, err := runner.IsBranchMerged(ctx, "feature", "main")
	if err != nil {
		t.Errorf("dry-run IsBranchMerged returned error: %v", err)
	}
	if !merged {
		t.Errorf("dry-run IsBranchMerged should return true")
	}

	// FetchAll
	if err := runner.FetchAll(ctx); err != nil {
		t.Errorf("dry-run FetchAll returned error: %v", err)
	}

	// WorktreePrune
	if err := runner.WorktreePrune(ctx); err != nil {
		t.Errorf("dry-run WorktreePrune returned error: %v", err)
	}

	// GetLastCommitAge
	age, err := runner.GetLastCommitAge(ctx, "/tmp/wt")
	if err != nil {
		t.Errorf("dry-run GetLastCommitAge returned error: %v", err)
	}
	if age != "unknown" {
		t.Errorf("dry-run GetLastCommitAge = %q, want %q", age, "unknown")
	}

	// GetBehindCount
	behind, err := runner.GetBehindCount(ctx, "/tmp/wt")
	if err != nil {
		t.Errorf("dry-run GetBehindCount returned error: %v", err)
	}
	if behind != 0 {
		t.Errorf("dry-run GetBehindCount = %d, want 0", behind)
	}

	// Pull
	if err := runner.Pull(ctx, "/tmp/wt"); err != nil {
		t.Errorf("dry-run Pull returned error: %v", err)
	}

	// PullRebase
	if err := runner.PullRebase(ctx, "/tmp/wt"); err != nil {
		t.Errorf("dry-run PullRebase returned error: %v", err)
	}
}

func TestParseDirtyStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{
			name:   "empty string",
			input:  "",
			expect: false,
		},
		{
			name:   "whitespace only",
			input:  "   \n\t\n  ",
			expect: false,
		},
		{
			name:   "single modified file",
			input:  " M cmd/root.go\n",
			expect: true,
		},
		{
			name:   "multiple files",
			input:  " M cmd/root.go\n?? newfile.txt\nA  added.go\n",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDirtyStatus(tt.input)
			if got != tt.expect {
				t.Errorf("parseDirtyStatus(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestParseDefaultBranch(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "full ref",
			input:  "refs/remotes/origin/main",
			expect: "main",
		},
		{
			name:   "with trailing newline",
			input:  "refs/remotes/origin/develop\n",
			expect: "develop",
		},
		{
			name:   "already short",
			input:  "main",
			expect: "main",
		},
		{
			name:   "empty",
			input:  "",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDefaultBranch(tt.input)
			if got != tt.expect {
				t.Errorf("parseDefaultBranch(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestParseBranchList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "typical output",
			input:  "  main\n  develop\n  feature/login",
			expect: []string{"main", "develop", "feature/login"},
		},
		{
			name:   "with current branch marker",
			input:  "* main\n  develop\n  feature/login",
			expect: []string{"main", "develop", "feature/login"},
		},
		{
			name:   "empty",
			input:  "",
			expect: nil,
		},
		{
			name:   "single branch",
			input:  "* main",
			expect: []string{"main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBranchList(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("parseBranchList(%q) got %v, want %v", tt.input, got, tt.expect)
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("branch[%d] = %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestParseBehindCount(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "zero",
			input:  "0\n",
			expect: 0,
		},
		{
			name:   "positive",
			input:  "5\n",
			expect: 5,
		},
		{
			name:   "invalid",
			input:  "not a number",
			expect: 0,
		},
		{
			name:   "empty",
			input:  "",
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBehindCount(tt.input)
			if got != tt.expect {
				t.Errorf("parseBehindCount(%q) = %d, want %d", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIntegrationCloneAndWorktree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Check git is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a source repo to clone from
	srcDir := t.TempDir()
	cmds := [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "config", "user.email", "test@test.com"},
		{"git", "-C", srcDir, "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %v\n%s", args, err, out)
		}
	}

	// Create a file and commit
	if err := os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	cmds = [][]string{
		{"git", "-C", srcDir, "add", "."},
		{"git", "-C", srcDir, "commit", "-m", "initial"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %v\n%s", args, err, out)
		}
	}

	// Clone bare
	projectDir := t.TempDir()
	bareDir := filepath.Join(projectDir, ".bare")

	runner := NewRunner(bareDir, false)
	ctx := context.Background()

	if err := runner.CloneBare(ctx, srcDir, bareDir); err != nil {
		t.Fatalf("CloneBare: %v", err)
	}

	// Configure remote fetch
	if err := runner.ConfigureRemoteFetch(ctx); err != nil {
		t.Fatalf("ConfigureRemoteFetch: %v", err)
	}

	// Fetch
	if err := runner.Fetch(ctx, "origin"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	// List remote branches
	branches, err := runner.ListRemoteBranches(ctx)
	if err != nil {
		t.Fatalf("ListRemoteBranches: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "main" || b == "master" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main or master in branches: %v", branches)
	}

	// Add a worktree
	defaultBranch := branches[0]
	wtPath := filepath.Join(projectDir, "worktrees", defaultBranch)
	if err := runner.WorktreeAdd(ctx, wtPath, defaultBranch); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	// Verify the worktree directory exists
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// List worktrees
	worktrees, err := runner.WorktreeList(ctx)
	if err != nil {
		t.Fatalf("WorktreeList: %v", err)
	}

	if len(worktrees) < 2 {
		t.Errorf("expected at least 2 worktrees (bare + added), got %d", len(worktrees))
	}
}
