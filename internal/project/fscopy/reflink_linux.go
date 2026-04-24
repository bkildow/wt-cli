//go:build linux

package fscopy

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// tryReflink clones src to dst via the Linux FICLONE ioctl, which is
// supported on btrfs and on XFS filesystems formatted with reflink=1
// (default since kernel 5.1 / xfsprogs 5.0). It fails with a mapped
// "unsupported" errno on ext4, NFS, tmpfs, cross-mount attempts, etc.;
// CopyFile catches that and falls back to a byte copy.
func tryReflink(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcF.Close() }()

	srcInfo, err := srcF.Stat()
	if err != nil {
		return err
	}

	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, srcInfo.Mode())
	if err != nil {
		return err
	}

	ioctlErr := unix.IoctlSetInt(int(dstF.Fd()), unix.FICLONE, int(srcF.Fd()))
	if closeErr := dstF.Close(); closeErr != nil && ioctlErr == nil {
		ioctlErr = closeErr
	}
	if ioctlErr != nil {
		_ = os.Remove(dst)
		if isUnsupported(ioctlErr) {
			return errReflinkUnsupported
		}
		return ioctlErr
	}
	return nil
}

// tryCloneTree is not supported on Linux: FICLONE is file-only and there
// is no tree-level reflink syscall on btrfs/XFS. Callers fall back to a
// per-file walk.
func tryCloneTree(_, _ string) error {
	return errReflinkUnsupported
}

func isUnsupported(err error) bool {
	var errno unix.Errno
	if !errors.As(err, &errno) {
		return false
	}
	// ENOTSUP and EOPNOTSUPP share the same underlying value on Linux, so
	// only one can appear as a switch case.
	switch errno {
	case unix.EOPNOTSUPP,
		unix.EXDEV,
		unix.ENOTTY,
		unix.EINVAL,
		unix.EBADF,
		unix.EPERM,
		unix.EISDIR,
		unix.ENOSYS:
		return true
	}
	return false
}
