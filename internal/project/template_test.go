package project

import (
	"testing"
)

func TestWorktreeIDFromBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"main", "main"},
		{"feature/login", "feature-login"},
		{"feature/auth/oauth", "feature-auth-oauth"},
		{"no-slashes", "no-slashes"},
		{"Feature/Auth", "feature-auth"},
	}
	for _, tt := range tests {
		got := WorktreeIDFromBranch(tt.branch)
		if got != tt.want {
			t.Errorf("WorktreeIDFromBranch(%q) = %q, want %q", tt.branch, got, tt.want)
		}
	}
}

func TestNewTemplateVars(t *testing.T) {
	vars := NewTemplateVars("/path/to/worktree", "feature/login")
	if vars.WorktreeID != "feature-login" {
		t.Errorf("WorktreeID = %q, want %q", vars.WorktreeID, "feature-login")
	}
	if vars.WorktreePath != "/path/to/worktree" {
		t.Errorf("WorktreePath = %q, want %q", vars.WorktreePath, "/path/to/worktree")
	}
	if vars.BranchName != "feature/login" {
		t.Errorf("BranchName = %q, want %q", vars.BranchName, "feature/login")
	}
}

func TestProcessTemplate(t *testing.T) {
	vars := TemplateVars{
		WorktreeID:   "feature-login",
		WorktreePath: "/project/worktrees/feature-login",
		BranchName:   "feature/login",
	}

	tests := []struct {
		input string
		want  string
	}{
		{"id: ${WORKTREE_ID}", "id: feature-login"},
		{"path: ${WORKTREE_PATH}", "path: /project/worktrees/feature-login"},
		{"branch: ${BRANCH_NAME}", "branch: feature/login"},
		{"${WORKTREE_ID}-app", "feature-login-app"},
		{"no placeholders", "no placeholders"},
	}
	for _, tt := range tests {
		got := ProcessTemplate(tt.input, vars)
		if got != tt.want {
			t.Errorf("ProcessTemplate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsTemplateFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".env.template", true},
		{"config.yml.template", true},
		{"plain.txt", false},
		{"template", false},
		{"foo.template.bak", false},
	}
	for _, tt := range tests {
		got := IsTemplateFile(tt.filename)
		if got != tt.want {
			t.Errorf("IsTemplateFile(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}

func TestStripTemplateExt(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{".env.template", ".env"},
		{"config.yml.template", "config.yml"},
		{"plain.txt", "plain.txt"},
	}
	for _, tt := range tests {
		got := StripTemplateExt(tt.filename)
		if got != tt.want {
			t.Errorf("StripTemplateExt(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}
