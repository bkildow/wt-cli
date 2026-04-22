//go:build darwin

package fscopy

import (
	"errors"

	"golang.org/x/sys/unix"
)

// tryReflink clones src to dst via APFS's clonefile(2). dst must not
// already exist; CopyFile guarantees that by writing to a fresh
// "<dst>.wtclone" path. Clonefile with flag 0 preserves mode, xattrs,
// and ACLs from src.
func tryReflink(src, dst string) error {
	err := unix.Clonefile(src, dst, 0)
	if err == nil {
		return nil
	}
	if isUnsupported(err) {
		return errReflinkUnsupported
	}
	return err
}

// isUnsupported reports whether err from Clonefile indicates that the
// call cannot succeed on this source/destination pair for reasons that
// are recoverable by falling back to a byte copy (wrong filesystem,
// cross-volume, permission mismatch on metadata copy, etc.) rather than
// a genuine I/O or logic error.
func isUnsupported(err error) bool {
	var errno unix.Errno
	if !errors.As(err, &errno) {
		return false
	}
	switch errno {
	case unix.ENOTSUP,
		unix.EOPNOTSUPP,
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
