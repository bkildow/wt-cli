//go:build !darwin && !linux

package fscopy

// tryReflink on platforms without a supported COW clone syscall (Windows,
// BSDs, etc.) always reports unsupported so CopyFile falls back to a
// byte copy. Windows ReFS does support block cloning via
// FSCTL_DUPLICATE_EXTENTS_TO_FILE, but ReFS is uncommon in dev
// environments, so we don't wire it up here.
func tryReflink(src, dst string) error {
	return errReflinkUnsupported
}

// tryCloneTree on platforms without a supported tree-clone syscall always
// reports unsupported so CopyTree's caller falls back to a per-file walk via
// IsReflinkUnsupported.
func tryCloneTree(src, dst string) error {
	return errReflinkUnsupported
}
