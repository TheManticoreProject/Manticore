//go:build !windows && !linux && !darwin && !freebsd && !dragonfly && !aix

package server

import "errors"

// errNoDiskUsage reports that the host offers no volume capacity query.
var errNoDiskUsage = errors.New("volume capacity is not available on this platform")

// diskUsage has no host query on this platform: js/wasm, wasip1 and plan9
// have no statfs, and openbsd, netbsd, solaris and illumos expose a statfs or
// statvfs layout the Unix implementation does not read. It returns an error
// so VolumeInfo leaves TotalBytes and FreeBytes at zero.
func diskUsage(path string) (total, free int64, err error) {
	return 0, 0, errNoDiskUsage
}

// hostVolumeSerial has no host query on this platform. ok is false so the
// caller falls back to the path-derived serial.
func hostVolumeSerial(path string) (serial uint32, ok bool) {
	return 0, false
}
