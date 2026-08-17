//go:build !windows

package disk

import "golang.org/x/sys/unix"

func stat(path string) (Usage, error) {
	var fs unix.Statfs_t
	if err := unix.Statfs(path, &fs); err != nil {
		return Usage{}, err
	}

	blockSize := uint64(fs.Bsize)
	return Usage{
		TotalBytes: fs.Blocks * blockSize,
		FreeBytes:  fs.Bavail * blockSize,
	}, nil
}
