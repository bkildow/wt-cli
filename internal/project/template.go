package project

import (
	"os"
	"strings"
)

type TemplateVars struct {
	WorktreeName string
	WorktreePath string
	BranchName   string
	DatabaseName string
}

func NewTemplateVars(worktreePath, branchName string) TemplateVars {
	name := WorktreeNameFromBranch(branchName)
	return TemplateVars{
		WorktreeName: name,
		WorktreePath: worktreePath,
		BranchName:   branchName,
		DatabaseName: SanitizeDatabaseName(name),
	}
}

func WorktreeNameFromBranch(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func SanitizeDatabaseName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return s
}

func ProcessTemplate(content string, vars TemplateVars) string {
	s := strings.ReplaceAll(content, "${WORKTREE_NAME}", vars.WorktreeName)
	s = strings.ReplaceAll(s, "${WORKTREE_PATH}", vars.WorktreePath)
	s = strings.ReplaceAll(s, "${BRANCH_NAME}", vars.BranchName)
	s = strings.ReplaceAll(s, "${DATABASE_NAME}", vars.DatabaseName)
	return s
}

func HasTemplateVars(content string) bool {
	return strings.Contains(content, "${WORKTREE_NAME}") ||
		strings.Contains(content, "${WORKTREE_PATH}") ||
		strings.Contains(content, "${BRANCH_NAME}") ||
		strings.Contains(content, "${DATABASE_NAME}")
}

func isBinaryFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}
