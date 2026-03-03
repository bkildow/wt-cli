package cmd

import (
	"testing"

	"github.com/bkildow/wt-cli/internal/git"
)

func TestFilterManagedWorktrees(t *testing.T) {
	tests := []struct {
		name        string
		worktrees   []git.WorktreeInfo
		projectRoot string
		wantCount   int
		wantBranch  []string
	}{
		{
			name: "bare repo setup filters bare entry",
			worktrees: []git.WorktreeInfo{
				{Path: "/proj/.bare", Branch: "", Bare: true},
				{Path: "/proj/worktrees/develop", Branch: "develop"},
			},
			projectRoot: "/proj",
			wantCount:   1,
			wantBranch:  []string{"develop"},
		},
		{
			name: "non-bare setup filters project root",
			worktrees: []git.WorktreeInfo{
				{Path: "/proj", Branch: "main"},
				{Path: "/proj/worktrees/feature", Branch: "feature"},
			},
			projectRoot: "/proj",
			wantCount:   1,
			wantBranch:  []string{"feature"},
		},
		{
			name: "non-bare setup with no additional worktrees",
			worktrees: []git.WorktreeInfo{
				{Path: "/proj", Branch: "main"},
			},
			projectRoot: "/proj",
			wantCount:   0,
		},
		{
			name: "multiple managed worktrees",
			worktrees: []git.WorktreeInfo{
				{Path: "/proj", Branch: "main"},
				{Path: "/proj/worktrees/feat-a", Branch: "feat-a"},
				{Path: "/proj/worktrees/feat-b", Branch: "feat-b"},
			},
			projectRoot: "/proj",
			wantCount:   2,
			wantBranch:  []string{"feat-a", "feat-b"},
		},
		{
			name:        "empty worktree list",
			worktrees:   nil,
			projectRoot: "/proj",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterManagedWorktrees(tt.worktrees, tt.projectRoot)
			if len(got) != tt.wantCount {
				t.Errorf("got %d worktrees, want %d", len(got), tt.wantCount)
			}
			for i, branch := range tt.wantBranch {
				if i >= len(got) {
					break
				}
				if got[i].Branch != branch {
					t.Errorf("got[%d].Branch = %q, want %q", i, got[i].Branch, branch)
				}
			}
		})
	}
}
