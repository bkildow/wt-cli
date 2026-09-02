// Package forge provides optional pull-request awareness backed by external
// CLIs. Everything here degrades gracefully: when no forge is usable, Detect
// returns nil and callers simply go without PR information.
package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// PRState mirrors the states a forge reports for a pull request.
type PRState string

const (
	StateOpen   PRState = "OPEN"
	StateMerged PRState = "MERGED"
	StateClosed PRState = "CLOSED"
)

// PullRequest is the minimal slice of PR data wt needs today. HeadOID is the
// commit the forge last saw on the head branch and BaseRef is what it merged
// into: checking both against local state is what stops a reused branch name,
// a post-merge push, or a PR against some other base from looking like merged
// work.
type PullRequest struct {
	Number  int
	State   PRState
	HeadOID string
	BaseRef string
}

// Forge looks up pull requests for a repository.
type Forge interface {
	// PRForBranch returns the most recent pull request whose head is branch,
	// or nil when the branch has none.
	PRForBranch(ctx context.Context, branch string) (*PullRequest, error)
}

// Detect returns a Forge for the given remote URL, or nil when none is usable:
// the gh CLI is not on PATH, or the remote is not a GitHub repository. A nil
// return is the normal "no forge here" answer, not an error.
func Detect(_ context.Context, remoteURL string) Forge {
	repo := parseGitHubRepo(remoteURL)
	if repo == "" {
		return nil
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return nil
	}
	return &GitHub{Repo: repo}
}

// GitHub implements Forge by shelling out to the gh CLI.
type GitHub struct {
	Repo string // owner/name
}

func (g *GitHub) PRForBranch(ctx context.Context, branch string) (*PullRequest, error) {
	args := []string{
		"pr", "list",
		"--repo", g.Repo,
		"--head", branch,
		"--state", "all",
		"--json", "number,state,headRefOid,baseRefName",
		"--limit", "1",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return parsePRList(stdout.Bytes())
}

// parsePRList decodes the JSON array `gh pr list --json number,state` emits.
// An empty array means the branch has no pull request.
func parsePRList(data []byte) (*PullRequest, error) {
	var prs []struct {
		Number  int    `json:"number"`
		State   string `json:"state"`
		HeadOID string `json:"headRefOid"`
		BaseRef string `json:"baseRefName"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(data), &prs); err != nil {
		return nil, fmt.Errorf("parsing gh output: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	return &PullRequest{
		Number:  prs[0].Number,
		State:   PRState(strings.ToUpper(prs[0].State)),
		HeadOID: prs[0].HeadOID,
		BaseRef: prs[0].BaseRef,
	}, nil
}

var githubRemote = regexp.MustCompile(`^(?:(?:https?|ssh|git)://(?:[^@/]+@)?github\.com/|(?:[^@]+@)?github\.com:)([^/]+/[^/]+?)(?:\.git)?/?$`)

// parseGitHubRepo extracts "owner/name" from a GitHub remote URL in any of the
// forms git supports, and returns "" for any other host.
func parseGitHubRepo(remoteURL string) string {
	m := githubRemote.FindStringSubmatch(strings.TrimSpace(remoteURL))
	if m == nil {
		return ""
	}
	return m[1]
}
