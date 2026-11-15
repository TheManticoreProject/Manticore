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
