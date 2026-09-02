package cmd

import (
	"context"
	"testing"

	"github.com/bkildow/wt-cli/internal/forge"
	"github.com/bkildow/wt-cli/internal/git"
)

// TestPRMerged pins the guards that stop a merged PR from condemning a
// worktree that is still live: gh matches PRs by head branch name alone.
func TestPRMerged(t *testing.T) {
	const head = "abc123"
	worktree := git.WorktreeInfo{Branch: "feature", Head: head}

	tests := []struct {
		name string
		pr   *forge.PullRequest
		want bool
	}{
		{
			"merged into the default branch at this tip",
			&forge.PullRequest{Number: 1, State: forge.StateMerged, HeadOID: head, BaseRef: "main"},
			true,
		},
		{"no pull request", nil, false},
		{
			"still open",
			&forge.PullRequest{Number: 1, State: forge.StateOpen, HeadOID: head, BaseRef: "main"},
			false,
		},
		{
			"closed without merging",
			&forge.PullRequest{Number: 1, State: forge.StateClosed, HeadOID: head, BaseRef: "main"},
			false,
		},
		{
			"merged into some other base",
			&forge.PullRequest{Number: 1, State: forge.StateMerged, HeadOID: head, BaseRef: "release/2.0"},
			false,
		},
		{
			"branch name reused, or commits pushed after the merge",
			&forge.PullRequest{Number: 1, State: forge.StateMerged, HeadOID: "999fff", BaseRef: "main"},
			false,
		},
		{
			"forge reported no head commit",
			&forge.PullRequest{Number: 1, State: forge.StateMerged, HeadOID: "", BaseRef: "main"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The worktree carries its head, so no git call is needed.
			if got := prMerged(context.Background(), nil, tt.pr, worktree, "main"); got != tt.want {
				t.Errorf("prMerged = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeReason(t *testing.T) {
	tests := []struct {
		method git.MergeMethod
		want   string
	}{
		{git.MergeAncestor, "merged"},
		{git.MergeRebase, "merged (rebase)"},
		{git.MergeSquash, "merged (squash)"},
		{git.MergeEquivalent, "merged (patch-equivalent)"},
		{git.MergeNone, "merged"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := mergeReason(tt.method); got != tt.want {
				t.Errorf("mergeReason(%q) = %q, want %q", tt.method, got, tt.want)
			}
		})
	}
}
