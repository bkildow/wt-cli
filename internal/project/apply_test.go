package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bkildow/wt-cli/internal/config"
)

func TestApplyCopy(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	// Create shared/copy/ with files
	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(filepath.Join(copyDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "sub", "nested.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplyCopy(root, wt, cfg, false, nil); err != nil {
		t.Fatalf("ApplyCopy error: %v", err)
	}

	// Verify files exist with correct content
	got, err := os.ReadFile(filepath.Join(wt, "file.txt"))
	if err != nil {
		t.Fatalf("file.txt not copied: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("file.txt content = %q, want %q", got, "hello")
	}

	got, err = os.ReadFile(filepath.Join(wt, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("sub/nested.txt not copied: %v", err)
	}
	if string(got) != "nested" {
		t.Errorf("sub/nested.txt content = %q, want %q", got, "nested")
	}
}

func TestApplyCopyDryRun(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplyCopy(root, wt, cfg, true, nil); err != nil {
		t.Fatalf("ApplyCopy dry-run error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(wt, "file.txt")); err == nil {
		t.Error("dry-run should not create file.txt")
	}
}

func TestApplyCopyMissingSharedDir(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplyCopy(root, wt, cfg, false, nil); err != nil {
		t.Fatalf("ApplyCopy with missing dir should not error: %v", err)
	}
}

func TestApplySymlinks(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	symlinkDir := filepath.Join(root, "shared", "symlink")
	nodeModules := filepath.Join(symlinkDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplySymlinks(root, wt, cfg, false); err != nil {
		t.Fatalf("ApplySymlinks error: %v", err)
	}

	link := filepath.Join(wt, "node_modules")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if target != nodeModules {
		t.Errorf("symlink target = %q, want %q", target, nodeModules)
	}
}

func TestApplySymlinksDryRun(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	symlinkDir := filepath.Join(root, "shared", "symlink")
	if err := os.MkdirAll(filepath.Join(symlinkDir, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplySymlinks(root, wt, cfg, true); err != nil {
		t.Fatalf("ApplySymlinks dry-run error: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(wt, "node_modules")); err == nil {
		t.Error("dry-run should not create symlink")
	}
}

func TestApplySymlinksMissingDir(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := ApplySymlinks(root, wt, cfg, false); err != nil {
		t.Fatalf("ApplySymlinks with missing dir should not error: %v", err)
	}
}

func TestApply(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	// Set up copy source
	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up symlink source
	symlinkDir := filepath.Join(root, "shared", "symlink")
	if err := os.MkdirAll(filepath.Join(symlinkDir, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	if _, err := Apply(root, wt, cfg, false, nil); err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	// Verify copy
	if _, err := os.Stat(filepath.Join(wt, "config.json")); err != nil {
		t.Error("Apply did not copy config.json")
	}

	// Verify symlink
	if _, err := os.Lstat(filepath.Join(wt, "vendor")); err != nil {
		t.Error("Apply did not create vendor symlink")
	}
}

func TestApplyCopyWithTemplateVars(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "ID=${WORKTREE_ID}\nPATH=${WORKTREE_PATH}\nROOT=${PROJECT_ROOT}\n"
	if err := os.WriteFile(filepath.Join(copyDir, ".env.template"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	vars := NewTemplateVars(root, wt, "feature/login")
	if _, err := ApplyCopy(root, wt, cfg, false, &vars); err != nil {
		t.Fatalf("ApplyCopy with vars error: %v", err)
	}

	// .template suffix should be stripped
	if _, err := os.Stat(filepath.Join(wt, ".env.template")); err == nil {
		t.Error(".env.template should not exist in output")
	}
	got, err := os.ReadFile(filepath.Join(wt, ".env"))
	if err != nil {
		t.Fatalf(".env not created: %v", err)
	}
	want := "ID=feature-login\nPATH=" + wt + "\nROOT=" + root + "\n"
	if string(got) != want {
		t.Errorf(".env content = %q, want %q", got, want)
	}
}

func TestApplySymlinksDirectoryConflict(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	// Source: shared/symlink/.claude/settings.local.json
	symlinkDir := filepath.Join(root, "shared", "symlink", ".claude")
	if err := os.MkdirAll(symlinkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsSrc := filepath.Join(symlinkDir, "settings.local.json")
	if err := os.WriteFile(settingsSrc, []byte(`{"hooks":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Destination: worktree/.claude/ already exists with its own file
	claudeDir := filepath.Join(wt, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existingFile := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(existingFile, []byte("# existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	count, err := ApplySymlinks(root, wt, cfg, false)
	if err != nil {
		t.Fatalf("ApplySymlinks error: %v", err)
	}
	if count != 1 {
		t.Errorf("symlink count = %d, want 1", count)
	}

	// settings.local.json should be a symlink to the shared source.
	link := filepath.Join(claudeDir, "settings.local.json")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if target != settingsSrc {
		t.Errorf("symlink target = %q, want %q", target, settingsSrc)
	}

	// Pre-existing file should be untouched.
	if _, err := os.Stat(existingFile); err != nil {
		t.Errorf("existing file was removed: %v", err)
	}
}

func TestApplyCopyNonTemplateFilesCopiedAsIs(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Non-.template files should be copied verbatim, even with ${...} patterns
	content := "value: ${WORKTREE_ID}\n"
	if err := os.WriteFile(filepath.Join(copyDir, "config.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{SharedDir: config.DefaultSharedDir}
	vars := NewTemplateVars(root, wt, "feature/login")
	if _, err := ApplyCopy(root, wt, cfg, false, &vars); err != nil {
		t.Fatalf("ApplyCopy error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wt, "config.yml"))
	if err != nil {
		t.Fatalf("config.yml not created: %v", err)
	}
	if string(got) != content {
		t.Errorf("non-template file was modified: got %q, want %q", got, content)
	}
}
