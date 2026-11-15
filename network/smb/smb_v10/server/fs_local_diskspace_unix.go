//go:build !windows

package server

import "golang.org/x/sys/unix"

// diskUsage returns the total and available capacity, in bytes, of the volume
// backing path. Available space is the space usable by an unprivileged caller
// (Bavail), matching what Windows reports to a client. A statfs failure is
// returned so the caller can fall back rather than report a bogus size.
func diskUsage(path string) (total, free int64, err error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	bs := int64(st.Bsize)
	return int64(st.Blocks) * bs, int64(st.Bavail) * bs, nil
}

// hostVolumeSerial returns a 32-bit serial for the volume backing path, derived
// from the filesystem id (statfs fsid), so the serial identifies the real host
// volume rather than the share path. ok is false when the fsid is unavailable or
// zero, so the caller can fall back.
func hostVolumeSerial(path string) (serial uint32, ok bool) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return 0, false
	}
	serial = uint32(st.Fsid.Val[0]) ^ uint32(st.Fsid.Val[1])
	if serial == 0 {
		return 0, false
	}
	return serial, true
}
