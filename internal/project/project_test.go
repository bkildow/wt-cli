package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bkildow/wt-cli/internal/config"
)

func TestRepoNameFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:org/repo.git", "repo"},
		{"git@github.com:org/repo", "repo"},
		{"https://github.com/org/repo.git", "repo"},
		{"https://github.com/org/repo", "repo"},
		{"git@gitlab.com:group/subgroup/repo.git", "repo"},
		{"https://gitlab.com/group/subgroup/repo.git", "repo"},
		{"git@github.com:user/my-project.git", "my-project"},
		{"https://github.com/user/my-project", "my-project"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := RepoNameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("RepoNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestFindRoot(t *testing.T) {
	// Create a nested directory structure with config at the root
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write config file at root
	cfg := config.DefaultConfig()
	if err := cfg.Save(root); err != nil {
		t.Fatal(err)
	}

	// FindRoot from nested dir should find root
	found, err := FindRoot(nested)
	if err != nil {
		t.Fatalf("FindRoot error: %v", err)
	}
	if found != root {
		t.Errorf("FindRoot = %q, want %q", found, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindRoot(dir)
	if !errors.Is(err, config.ErrConfigNotFound) {
		t.Errorf("err = %v, want ErrConfigNotFound", err)
	}
}

func TestCreateScaffold(t *testing.T) {
	root := t.TempDir()

	if err := CreateScaffold(root, false); err != nil {
		t.Fatalf("CreateScaffold error: %v", err)
	}

	dirs := []string{
		filepath.Join(root, "shared", "copy"),
		filepath.Join(root, "shared", "symlink"),
		filepath.Join(root, "worktrees"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %q not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

func TestCreateScaffoldDryRun(t *testing.T) {
	root := t.TempDir()

	if err := CreateScaffold(root, true); err != nil {
		t.Fatalf("CreateScaffold dry-run error: %v", err)
	}

	// In dry-run, directories should NOT be created
	dirs := []string{
		filepath.Join(root, "shared"),
		filepath.Join(root, "worktrees"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			t.Errorf("dry-run should not create %q", dir)
		}
	}
}

func TestGitDirPath(t *testing.T) {
	cfg := &config.Config{GitDir: ".bare"}
	root := "/home/user/project"
	got := GitDirPath(root, cfg)
	want := filepath.Join(root, ".bare")
	if got != want {
		t.Errorf("GitDirPath = %q, want %q", got, want)
	}
}
