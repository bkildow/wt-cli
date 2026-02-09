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
setup:
  - npm install
teardown:
  - "docker compose down"
editor: cursor
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
	if len(cfg.Setup) != 1 || cfg.Setup[0] != "npm install" {
		t.Errorf("setup = %v, want [npm install]", cfg.Setup)
	}
	if len(cfg.Teardown) != 1 || cfg.Teardown[0] != "docker compose down" {
		t.Errorf("teardown = %v, want [docker compose down]", cfg.Teardown)
	}
	if cfg.Editor != "cursor" {
		t.Errorf("editor = %q, want %q", cfg.Editor, "cursor")
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
		Version:  1,
		GitDir:   ".bare",
		Setup:    []string{"make build", "make test"},
		Teardown: []string{"make clean"},
		Editor:   "nvim",
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
	if len(loaded.Setup) != len(original.Setup) {
		t.Errorf("setup len = %d, want %d", len(loaded.Setup), len(original.Setup))
	}
	if len(loaded.Teardown) != len(original.Teardown) {
		t.Errorf("teardown len = %d, want %d", len(loaded.Teardown), len(original.Teardown))
	}
	if loaded.Teardown[0] != "make clean" {
		t.Errorf("teardown[0] = %q, want %q", loaded.Teardown[0], "make clean")
	}
	if loaded.Editor != original.Editor {
		t.Errorf("editor = %q, want %q", loaded.Editor, original.Editor)
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
