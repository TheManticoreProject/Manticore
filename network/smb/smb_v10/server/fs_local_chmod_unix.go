//go:build unix

package server

import (
	"errors"
	"os"
)

// setReadOnly sets or clears the owner's write permission through a handle
// opened beneath the root, so the file changed is the one the share names.
// Changing the mode needs only a handle, so a file the server may not read is
// opened for writing instead.
func (fs *LocalFileSystem) setReadOnly(path string, readOnly bool) error {
	name := rel(path)
	file, err := fs.openRegular(name, os.O_RDONLY, 0, true)
	if errors.Is(err, ErrAccessDenied) && !errors.Is(err, errNotRegular) {
		file, err = fs.openRegular(name, os.O_WRONLY, 0, false)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return translate(err)
	}
	perm := info.Mode().Perm()
	if readOnly {
		perm &= ^os.FileMode(0o222)
	} else {
		perm |= 0o200
	}
	return translate(file.Chmod(perm))
}
