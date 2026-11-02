//go:build !unix

package server

import "os"

// setReadOnly sets or clears the read-only attribute by path. On Windows a
// handle's Chmod needs FILE_WRITE_ATTRIBUTES, which a read-only open lacks and
// an open for writing of a read-only file is refused, so the path is resolved
// and checked against the root by hostPath, as before os.Root, until Go 1.25's
// Root.Chmod (#1896).
func (fs *LocalFileSystem) setReadOnly(path string, readOnly bool) error {
	host, err := fs.hostPath(path)
	if err != nil {
		return translate(err)
	}
	info, err := os.Stat(host)
	if err != nil {
		return translate(err)
	}
	perm := info.Mode().Perm()
	if readOnly {
		perm &= ^os.FileMode(0o222)
	} else {
		perm |= 0o200
	}
	return translate(os.Chmod(host, perm))
}
