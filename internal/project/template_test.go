package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorktreeNameFromBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"main", "main"},
		{"feature/login", "feature-login"},
		{"feature/auth/oauth", "feature-auth-oauth"},
		{"no-slashes", "no-slashes"},
	}
	for _, tt := range tests {
		got := WorktreeNameFromBranch(tt.branch)
		if got != tt.want {
			t.Errorf("WorktreeNameFromBranch(%q) = %q, want %q", tt.branch, got, tt.want)
		}
	}
}

func TestSanitizeDatabaseName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"feature-login", "feature_login"},
		{"Feature-Login", "feature_login"},
		{"my.project", "my_project"},
		{"my-app.v2", "my_app_v2"},
		{"already_clean", "already_clean"},
	}
	for _, tt := range tests {
		got := SanitizeDatabaseName(tt.name)
		if got != tt.want {
			t.Errorf("SanitizeDatabaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNewTemplateVars(t *testing.T) {
	vars := NewTemplateVars("/path/to/worktree", "feature/login")
	if vars.WorktreeName != "feature-login" {
		t.Errorf("WorktreeName = %q, want %q", vars.WorktreeName, "feature-login")
	}
	if vars.WorktreePath != "/path/to/worktree" {
		t.Errorf("WorktreePath = %q, want %q", vars.WorktreePath, "/path/to/worktree")
	}
	if vars.BranchName != "feature/login" {
		t.Errorf("BranchName = %q, want %q", vars.BranchName, "feature/login")
	}
	if vars.DatabaseName != "feature_login" {
		t.Errorf("DatabaseName = %q, want %q", vars.DatabaseName, "feature_login")
	}
}

func TestProcessTemplate(t *testing.T) {
	vars := TemplateVars{
		WorktreeName: "feature-login",
		WorktreePath: "/project/worktrees/feature-login",
		BranchName:   "feature/login",
		DatabaseName: "feature_login",
	}

	tests := []struct {
		input string
		want  string
	}{
		{"DB=${DATABASE_NAME}", "DB=feature_login"},
		{"name: ${WORKTREE_NAME}", "name: feature-login"},
		{"path: ${WORKTREE_PATH}", "path: /project/worktrees/feature-login"},
		{"branch: ${BRANCH_NAME}", "branch: feature/login"},
		{"${WORKTREE_NAME}-${DATABASE_NAME}", "feature-login-feature_login"},
		{"no placeholders", "no placeholders"},
	}
	for _, tt := range tests {
		got := ProcessTemplate(tt.input, vars)
		if got != tt.want {
			t.Errorf("ProcessTemplate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHasTemplateVars(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"${WORKTREE_NAME}", true},
		{"${WORKTREE_PATH}", true},
		{"${BRANCH_NAME}", true},
		{"${DATABASE_NAME}", true},
		{"prefix ${WORKTREE_NAME} suffix", true},
		{"no placeholders here", false},
		{"${UNKNOWN_VAR}", false},
	}
	for _, tt := range tests {
		got := HasTemplateVars(tt.content)
		if got != tt.want {
			t.Errorf("HasTemplateVars(%q) = %v, want %v", tt.content, got, tt.want)
		}
	}
}

func TestIsBinaryFile(t *testing.T) {
	dir := t.TempDir()

	textFile := filepath.Join(dir, "text.txt")
	if err := os.WriteFile(textFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := isBinaryFile(textFile)
	if err != nil {
		t.Fatalf("isBinaryFile(text) error: %v", err)
	}
	if got {
		t.Error("isBinaryFile(text) = true, want false")
	}

	binFile := filepath.Join(dir, "binary.bin")
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x0D, 0x0A}
	if err := os.WriteFile(binFile, data, 0644); err != nil {
		t.Fatal(err)
	}
	got, err = isBinaryFile(binFile)
	if err != nil {
		t.Fatalf("isBinaryFile(binary) error: %v", err)
	}
	if !got {
		t.Error("isBinaryFile(binary) = false, want true")
	}
}
