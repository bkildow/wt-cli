//go:build windows

package disk

import "golang.org/x/sys/windows"

func stat(path string) (Usage, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return Usage{}, err
	}

	var freeToCaller, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &freeToCaller, &total, &free); err != nil {
		return Usage{}, err
	}

	return Usage{
		TotalBytes: total,
		FreeBytes:  freeToCaller,
	}, nil
}
