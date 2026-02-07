package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	content := `version: 1
git_dir: .bare
post_create:
  - npm install
project_type: node
`
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if cfg.GitDir != ".bare" {
		t.Errorf("git_dir = %q, want %q", cfg.GitDir, ".bare")
	}
	if len(cfg.PostCreate) != 1 || cfg.PostCreate[0] != "npm install" {
		t.Errorf("post_create = %v, want [npm install]", cfg.PostCreate)
	}
	if cfg.ProjectType != "node" {
		t.Errorf("project_type = %q, want %q", cfg.ProjectType, "node")
	}
}

func TestLoadMinimalConfig(t *testing.T) {
	dir := t.TempDir()
	content := `version: 2
`
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("version = %d, want 2", cfg.Version)
	}
	// Default should be applied for git_dir
	if cfg.GitDir != DefaultGitDir {
		t.Errorf("git_dir = %q, want default %q", cfg.GitDir, DefaultGitDir)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err != ErrConfigNotFound {
		t.Errorf("err = %v, want ErrConfigNotFound", err)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := `{{{not yaml`
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	original := &Config{
		Version:     1,
		GitDir:      ".bare",
		PostCreate:  []string{"make build", "make test"},
		ProjectType: "go",
	}

	if err := original.Save(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version = %d, want %d", loaded.Version, original.Version)
	}
	if loaded.GitDir != original.GitDir {
		t.Errorf("git_dir = %q, want %q", loaded.GitDir, original.GitDir)
	}
	if len(loaded.PostCreate) != len(original.PostCreate) {
		t.Errorf("post_create len = %d, want %d", len(loaded.PostCreate), len(original.PostCreate))
	}
	if loaded.ProjectType != original.ProjectType {
		t.Errorf("project_type = %q, want %q", loaded.ProjectType, original.ProjectType)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("Exists should return false when config file is missing")
	}

	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !Exists(dir) {
		t.Error("Exists should return true when config file is present")
	}
}
