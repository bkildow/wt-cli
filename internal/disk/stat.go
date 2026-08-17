package disk

import "golang.org/x/sys/unix"

func stat(path string) (Usage, error) {
	var fs unix.Statfs_t
	if err := unix.Statfs(path, &fs); err != nil {
		return Usage{}, err
	}

	// Bsize is a filesystem block size, so it is never negative. Its type
	// varies by platform (int64 on Linux, uint32 on Darwin), which is why the
	// conversion needs silencing on one and not the other.
	blockSize := uint64(fs.Bsize) //nolint:gosec // G115: block size is never negative
	return Usage{
		TotalBytes: fs.Blocks * blockSize,
		FreeBytes:  fs.Bavail * blockSize,
	}, nil
}
