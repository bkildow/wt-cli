package forge

import (
	"context"
	"testing"
)

func TestParsePRList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *PullRequest
		wantErr bool
	}{
		{
			"merged",
			`[{"number":123,"state":"MERGED","headRefOid":"abc123","baseRefName":"main"}]`,
			&PullRequest{Number: 123, State: StateMerged, HeadOID: "abc123", BaseRef: "main"},
			false,
		},
		{
			"open",
			`[{"number":7,"state":"OPEN","headRefOid":"def456","baseRefName":"develop"}]`,
			&PullRequest{Number: 7, State: StateOpen, HeadOID: "def456", BaseRef: "develop"},
			false,
		},
		{
			"closed",
			`[{"number":9,"state":"CLOSED","headRefOid":"aaa","baseRefName":"main"}]`,
			&PullRequest{Number: 9, State: StateClosed, HeadOID: "aaa", BaseRef: "main"},
			false,
		},
		{
			"lowercase state",
			`[{"number":9,"state":"merged","headRefOid":"aaa","baseRefName":"main"}]`,
			&PullRequest{Number: 9, State: StateMerged, HeadOID: "aaa", BaseRef: "main"},
			false,
		},
		{"no pull request", `[]`, nil, false},
		{"trailing newline", "[]\n", nil, false},
		{"malformed", `{"number":1}`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr, err := parsePRList([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("parsePRList: expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePRList: %v", err)
			}
			if tt.want == nil {
				if pr != nil {
					t.Fatalf("parsePRList = %+v, want nil", pr)
				}
				return
			}
			if pr == nil {
				t.Fatal("parsePRList = nil, want a pull request")
			}
			if *pr != *tt.want {
				t.Errorf("parsePRList = %+v, want %+v", *pr, *tt.want)
			}
		})
	}
}

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git@github.com:bkildow/wt-cli.git", "bkildow/wt-cli"},
		{"git@github.com:bkildow/wt-cli", "bkildow/wt-cli"},
		{"https://github.com/bkildow/wt-cli.git", "bkildow/wt-cli"},
		{"https://github.com/bkildow/wt-cli", "bkildow/wt-cli"},
		{"https://user@github.com/bkildow/wt-cli.git", "bkildow/wt-cli"},
		{"ssh://git@github.com/bkildow/wt-cli.git", "bkildow/wt-cli"},
		{"  https://github.com/bkildow/wt-cli/  ", "bkildow/wt-cli"},
		{"git@gitlab.com:bkildow/wt-cli.git", ""},
		{"https://bitbucket.org/bkildow/wt-cli.git", ""},
		{"/local/path/to/repo.git", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseGitHubRepo(tt.input); got != tt.want {
				t.Errorf("parseGitHubRepo(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectNonGitHubRemoteReturnsNil(t *testing.T) {
	if f := Detect(context.Background(), "git@gitlab.com:bkildow/wt-cli.git"); f != nil {
		t.Errorf("Detect(gitlab remote) = %v, want nil", f)
	}
	if f := Detect(context.Background(), ""); f != nil {
		t.Errorf("Detect(empty remote) = %v, want nil", f)
	}
}
