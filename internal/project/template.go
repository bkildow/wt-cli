package project

import (
	"path/filepath"
	"strings"
)

type TemplateVars struct {
	ProjectRoot  string
	WorktreeID   string
	WorktreePath string
	BranchName   string
}

func NewTemplateVars(projectRoot, worktreePath, branchName string) TemplateVars {
	return TemplateVars{
		ProjectRoot:  filepath.Clean(projectRoot),
		WorktreeID:   WorktreeIDFromBranch(branchName),
		WorktreePath: worktreePath,
		BranchName:   branchName,
	}
}

func WorktreeIDFromBranch(branch string) string {
	return strings.ToLower(strings.ReplaceAll(branch, "/", "-"))
}

func ProcessTemplate(content string, vars TemplateVars) string {
	s := strings.ReplaceAll(content, "${PROJECT_ROOT}", vars.ProjectRoot)
	s = strings.ReplaceAll(s, "${WORKTREE_ID}", vars.WorktreeID)
	s = strings.ReplaceAll(s, "${WORKTREE_PATH}", vars.WorktreePath)
	s = strings.ReplaceAll(s, "${BRANCH_NAME}", vars.BranchName)
	return s
}

func IsTemplateFile(filename string) bool {
	return strings.HasSuffix(filename, ".template")
}

func StripTemplateExt(filename string) string {
	return strings.TrimSuffix(filename, ".template")
}
