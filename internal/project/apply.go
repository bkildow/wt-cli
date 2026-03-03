package project

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bkildow/wt-cli/internal/ui"
)

// ApplyResult holds counts of files processed during Apply.
type ApplyResult struct {
	Copied    int
	Symlinked int
}

func ApplyCopy(projectRoot, worktreePath string, dryRun bool, vars *TemplateVars) (int, error) {
	copyDir := filepath.Join(projectRoot, "shared", "copy")

	if _, err := os.Stat(copyDir); os.IsNotExist(err) {
		return 0, nil
	}

	ui.Step("Copying shared files")

	var count int
	logged := make(map[string]bool)
	sep := string(filepath.Separator)
	err := filepath.WalkDir(copyDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(copyDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(worktreePath, rel)

		// Determine the top-level entry for collapsed logging.
		topLevel, _, isNested := strings.Cut(rel, sep)

		if dryRun {
			if isNested {
				logCopyDir(topLevel, logged, true)
			} else {
				dryDest := dest
				if vars != nil && IsTemplateFile(rel) {
					dryDest = filepath.Join(worktreePath, StripTemplateExt(rel))
				}
				ui.DryRunNotice(fmt.Sprintf("copy %s -> %s", path, dryDest))
			}
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		if vars != nil && IsTemplateFile(rel) {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			srcInfo, err := os.Stat(path)
			if err != nil {
				return err
			}
			dest = filepath.Join(worktreePath, StripTemplateExt(rel))
			processed := ProcessTemplate(string(content), *vars)
			ui.Info(fmt.Sprintf("  substituted template variables in %s", StripTemplateExt(rel)))
			count++
			return os.WriteFile(dest, []byte(processed), srcInfo.Mode())
		}

		if err := copyFile(path, dest); err != nil {
			return err
		}
		if isNested {
			logCopyDir(topLevel, logged, false)
		} else {
			ui.Info(fmt.Sprintf("  copied %s", rel))
		}
		count++
		return nil
	})
	return count, err
}

// ApplySymlinks creates symlinks in the worktree for each top-level entry
// in shared/symlink/.
func ApplySymlinks(projectRoot, worktreePath string, dryRun bool) (int, error) {
	symlinkDir := filepath.Join(projectRoot, "shared", "symlink")

	if _, err := os.Stat(symlinkDir); os.IsNotExist(err) {
		return 0, nil
	}

	ui.Step("Creating symlinks")

	entries, err := os.ReadDir(symlinkDir)
	if err != nil {
		return 0, err
	}

	var count int
	for _, entry := range entries {
		target := filepath.Join(symlinkDir, entry.Name())
		link := filepath.Join(worktreePath, entry.Name())

		if dryRun {
			ui.DryRunNotice(fmt.Sprintf("symlink %s -> %s", link, target))
			continue
		}

		// Remove existing file/symlink at destination (error ignored; Symlink will fail if needed)
		_ = os.Remove(link)

		if err := os.Symlink(target, link); err != nil {
			return count, err
		}

		relTarget, _ := filepath.Rel(worktreePath, target)
		ui.Info(fmt.Sprintf("  symlinked %s → %s", entry.Name(), relTarget))
		count++
	}

	return count, nil
}

func Apply(projectRoot, worktreePath string, dryRun bool, vars *TemplateVars) (ApplyResult, error) {
	copied, err := ApplyCopy(projectRoot, worktreePath, dryRun, vars)
	if err != nil {
		return ApplyResult{}, err
	}
	symlinked, err := ApplySymlinks(projectRoot, worktreePath, dryRun)
	if err != nil {
		return ApplyResult{}, err
	}
	return ApplyResult{Copied: copied, Symlinked: symlinked}, nil
}

// logCopyDir logs a top-level directory copy once, collapsing nested files.
func logCopyDir(name string, logged map[string]bool, dryRun bool) {
	if logged[name] {
		return
	}
	if dryRun {
		ui.DryRunNotice(fmt.Sprintf("copy %s/", name))
	} else {
		ui.Info(fmt.Sprintf("  copied %s/", name))
	}
	logged[name] = true
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}
