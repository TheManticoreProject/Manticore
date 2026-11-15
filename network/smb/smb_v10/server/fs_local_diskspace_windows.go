//go:build windows

package server

import "golang.org/x/sys/windows"

// diskUsage returns the total and available capacity, in bytes, of the volume
// backing path. freeBytesAvailableToCaller is the space usable by the calling
// account, matching what a client expects. A failure is returned so the caller
// can fall back rather than report a bogus size.
func diskUsage(path string) (total, free int64, err error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}
	var freeToCaller, totalBytes, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &freeToCaller, &totalBytes, &totalFree); err != nil {
		return 0, 0, err
	}
	return int64(totalBytes), int64(freeToCaller), nil
}
