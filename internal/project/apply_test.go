package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCopy(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	// Create shared/copy/ with files
	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(filepath.Join(copyDir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "sub", "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ApplyCopy(root, wt, false, nil); err != nil {
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
	if err := os.MkdirAll(copyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ApplyCopy(root, wt, true, nil); err != nil {
		t.Fatalf("ApplyCopy dry-run error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(wt, "file.txt")); err == nil {
		t.Error("dry-run should not create file.txt")
	}
}

func TestApplyCopyMissingSharedDir(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	if err := ApplyCopy(root, wt, false, nil); err != nil {
		t.Fatalf("ApplyCopy with missing dir should not error: %v", err)
	}
}

func TestApplySymlinks(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	symlinkDir := filepath.Join(root, "shared", "symlink")
	nodeModules := filepath.Join(symlinkDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatal(err)
	}

	if err := ApplySymlinks(root, wt, false); err != nil {
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
	if err := os.MkdirAll(filepath.Join(symlinkDir, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ApplySymlinks(root, wt, true); err != nil {
		t.Fatalf("ApplySymlinks dry-run error: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(wt, "node_modules")); err == nil {
		t.Error("dry-run should not create symlink")
	}
}

func TestApplySymlinksMissingDir(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	if err := ApplySymlinks(root, wt, false); err != nil {
		t.Fatalf("ApplySymlinks with missing dir should not error: %v", err)
	}
}

func TestApply(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	// Set up copy source
	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up symlink source
	symlinkDir := filepath.Join(root, "shared", "symlink")
	if err := os.MkdirAll(filepath.Join(symlinkDir, "vendor"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := Apply(root, wt, false, nil); err != nil {
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
	if err := os.MkdirAll(copyDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "db: ${DATABASE_NAME}\nname: ${WORKTREE_NAME}\n"
	if err := os.WriteFile(filepath.Join(copyDir, "config.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	vars := NewTemplateVars(wt, "feature/login")
	if err := ApplyCopy(root, wt, false, &vars); err != nil {
		t.Fatalf("ApplyCopy with vars error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wt, "config.yml"))
	if err != nil {
		t.Fatalf("config.yml not created: %v", err)
	}
	want := "db: feature_login\nname: feature-login\n"
	if string(got) != want {
		t.Errorf("config.yml content = %q, want %q", got, want)
	}
}

func TestApplyCopyBinaryFileSkipsTemplate(t *testing.T) {
	root := t.TempDir()
	wt := t.TempDir()

	copyDir := filepath.Join(root, "shared", "copy")
	if err := os.MkdirAll(copyDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("${WORKTREE_NAME}")
	content = append(content, 0x00)
	if err := os.WriteFile(filepath.Join(copyDir, "image.bin"), content, 0644); err != nil {
		t.Fatal(err)
	}

	vars := NewTemplateVars(wt, "feature/login")
	if err := ApplyCopy(root, wt, false, &vars); err != nil {
		t.Fatalf("ApplyCopy binary error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wt, "image.bin"))
	if err != nil {
		t.Fatalf("image.bin not created: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("binary file was modified, want literal preservation")
	}
}
