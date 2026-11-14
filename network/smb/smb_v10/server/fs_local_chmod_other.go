//go:build !unix

package server

import "os"

// setReadOnly sets or clears the read-only attribute. On Windows a handle's Chmod
// needs FILE_WRITE_ATTRIBUTES, which a read-only open through openRegular lacks,
// so the attribute goes through Root.Chmod: it opens the file beneath the root
// with the access the change needs, so the file changed is the one the share
// names and a directory swapped for a link out of the share cannot redirect it
// (#1896).
func (fs *LocalFileSystem) setReadOnly(path string, readOnly bool) error {
	name := rel(path)
	info, err := fs.rootDir.Stat(name)
	if err != nil {
		return translate(err)
	}
	perm := info.Mode().Perm()
	if readOnly {
		perm &= ^os.FileMode(0o222)
	} else {
		perm |= 0o200
	}
	return translate(fs.rootDir.Chmod(name, perm))
}
