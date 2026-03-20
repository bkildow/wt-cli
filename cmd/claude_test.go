package cmd

import (
	"strings"
	"testing"
)

func TestReadHookPayload_valid(t *testing.T) {
	input := `{"session_id":"abc","worktree_name":"feature/test","project_dir":"/tmp/proj","cwd":"/tmp/proj"}`
	payload, err := readHookPayload(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.SessionID != "abc" {
		t.Errorf("session_id = %q, want %q", payload.SessionID, "abc")
	}
	if payload.WorktreeName != "feature/test" {
		t.Errorf("worktree_name = %q, want %q", payload.WorktreeName, "feature/test")
	}
	if payload.ProjectDir != "/tmp/proj" {
		t.Errorf("project_dir = %q, want %q", payload.ProjectDir, "/tmp/proj")
	}
}

func TestReadHookPayload_empty(t *testing.T) {
	_, err := readHookPayload(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if !strings.Contains(err.Error(), "no payload") {
		t.Errorf("error = %q, want 'no payload'", err.Error())
	}
}

func TestReadHookPayload_invalidJSON(t *testing.T) {
	_, err := readHookPayload(strings.NewReader("{invalid"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error = %q, want 'invalid JSON'", err.Error())
	}
}

func TestReadHookPayload_missingFields(t *testing.T) {
	// Missing fields should not error — they are just empty strings.
	input := `{"session_id":"abc"}`
	payload, err := readHookPayload(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WorktreeName != "" {
		t.Errorf("worktree_name should be empty, got %q", payload.WorktreeName)
	}
}

func TestReadHookPayload_extraFields(t *testing.T) {
	// Extra fields should be silently ignored.
	input := `{"worktree_name":"test","unknown_field":"value"}`
	payload, err := readHookPayload(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WorktreeName != "test" {
		t.Errorf("worktree_name = %q, want %q", payload.WorktreeName, "test")
	}
}

func TestResolveProjectRoot_noProject(t *testing.T) {
	payload := hookPayload{
		ProjectDir: "/nonexistent/path",
		Cwd:        "/also/nonexistent",
	}
	_, err := resolveProjectRoot(payload)
	if err == nil {
		t.Fatal("expected error for nonexistent paths")
	}
	if !strings.Contains(err.Error(), "could not find wt project root") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestResolveProjectRoot_emptyPayload(t *testing.T) {
	payload := hookPayload{}
	_, err := resolveProjectRoot(payload)
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}
