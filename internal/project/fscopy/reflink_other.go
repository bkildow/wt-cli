//go:build !darwin && !linux

package fscopy

// tryReflink on platforms without a supported COW clone syscall (the BSDs,
// etc.) always reports unsupported so CopyFile falls back to a byte copy.
func tryReflink(src, dst string) error {
	return errReflinkUnsupported
}

// tryCloneTree on platforms without a supported tree-clone syscall always
// reports unsupported so CopyTree's caller falls back to a per-file walk via
// IsReflinkUnsupported.
func tryCloneTree(src, dst string) error {
	return errReflinkUnsupported
}
