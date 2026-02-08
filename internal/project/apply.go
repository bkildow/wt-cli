package project

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/briankildow/wt-cli/internal/ui"
)

// ApplyCopy walks shared/copy/ and copies each file into the worktree,
// preserving directory structure and file permissions.
func ApplyCopy(projectRoot, worktreePath string, dryRun bool) error {
	copyDir := filepath.Join(projectRoot, "shared", "copy")

	if _, err := os.Stat(copyDir); os.IsNotExist(err) {
		return nil
	}

	ui.Step("Copying shared files")

	return filepath.WalkDir(copyDir, func(path string, d fs.DirEntry, err error) error {
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

		if dryRun {
			ui.DryRunNotice(fmt.Sprintf("copy %s -> %s", path, dest))
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		return copyFile(path, dest)
	})
}

// ApplySymlinks creates symlinks in the worktree for each top-level entry
// in shared/symlink/.
func ApplySymlinks(projectRoot, worktreePath string, dryRun bool) error {
	symlinkDir := filepath.Join(projectRoot, "shared", "symlink")

	if _, err := os.Stat(symlinkDir); os.IsNotExist(err) {
		return nil
	}

	ui.Step("Creating symlinks")

	entries, err := os.ReadDir(symlinkDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		target := filepath.Join(symlinkDir, entry.Name())
		link := filepath.Join(worktreePath, entry.Name())

		if dryRun {
			ui.DryRunNotice(fmt.Sprintf("symlink %s -> %s", link, target))
			continue
		}

		// Remove existing file/symlink at destination
		os.Remove(link)

		if err := os.Symlink(target, link); err != nil {
			return err
		}
	}

	return nil
}

// Apply runs ApplyCopy then ApplySymlinks, returning the first error.
func Apply(projectRoot, worktreePath string, dryRun bool) error {
	if err := ApplyCopy(projectRoot, worktreePath, dryRun); err != nil {
		return err
	}
	return ApplySymlinks(projectRoot, worktreePath, dryRun)
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
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
