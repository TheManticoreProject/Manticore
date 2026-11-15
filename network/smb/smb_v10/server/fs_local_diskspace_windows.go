//go:build windows

package server

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

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

// hostVolumeSerial returns the real volume serial of the volume backing path,
// via GetVolumeInformation on the volume root. ok is false on failure so the
// caller can fall back.
func hostVolumeSerial(path string) (serial uint32, ok bool) {
	root := filepath.VolumeName(path) + `\`
	p, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return 0, false
	}
	if err := windows.GetVolumeInformation(p, nil, 0, &serial, nil, nil, nil, 0); err != nil {
		return 0, false
	}
	return serial, true
}
