package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LocalFileSystem is a FileSystem backed by a directory on the host.
//
// The share is rooted at that directory and nothing above it is reachable. Two
// separate mechanisms keep that true, because either alone is insufficient:
// resolvePath refuses a path that could climb out, and every resolved path is
// checked against the root again after the host has followed any symbolic links
// in it. The first stops a traversal spelled in the request; the second stops one
// planted in the file system itself, which no amount of path checking can see.
type LocalFileSystem struct {
	// root is the absolute, symlink-resolved directory the share is rooted at.
	root string

	// label is the volume label reported to a client.
	label string

	// readOnly refuses every modifying operation at the backend, independently of
	// the share's own flag, so a caller can hand out a genuinely read-only
	// backend.
	readOnly bool
}

// NewLocalFileSystem roots a file system at a directory on the host.
//
// The directory is resolved through any symbolic links at construction, so the
// containment checks compare like with like afterwards.
//
// Parameters:
//   - root: the directory to serve
//   - label: the volume label reported to a client
//
// Returns:
//   - The file system
//   - An error if the directory is unusable
func NewLocalFileSystem(root, label string) (*LocalFileSystem, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %q: %v", root, err)
	}
	// Resolve the root's own symlinks once, so a later comparison is against the
	// real directory rather than a link to it.
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %q: %v", absolute, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %q: %v", resolved, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", resolved)
	}

	return &LocalFileSystem{root: resolved, label: label}, nil
}

// SetReadOnly refuses every modifying operation at the backend.
func (fs *LocalFileSystem) SetReadOnly(readOnly bool) {
	fs.readOnly = readOnly
}

// Root returns the directory the share is rooted at.
func (fs *LocalFileSystem) Root() string {
	return fs.root
}

// hostPath maps a resolved share-relative path to a host path, and refuses one
// that leaves the root.
//
// This is the second containment mechanism. resolvePath has already refused
// anything that spells an escape, but it cannot see a symbolic link inside the
// share pointing out of it: the path "public/link/secret" is perfectly well
// formed. So the host path is resolved through its links and compared against the
// root before it is used.
//
// A path that does not exist yet cannot be resolved through links, so its parent
// is checked instead — which is what a create needs, and is sufficient, because a
// link can only be traversed through a component that exists.
func (fs *LocalFileSystem) hostPath(path string) (string, error) {
	// filepath.Join cleans the result, but the input is already known to hold no
	// ".." element, so cleaning cannot move it upwards.
	candidate := filepath.Join(fs.root, filepath.FromSlash(path))

	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		// The target does not exist. Its parent must, and must be inside the
		// root; the final element cannot itself be a traversed link, because it
		// is not there to be one.
		parent := filepath.Dir(candidate)
		resolvedParent, parentErr := filepath.EvalSymlinks(parent)
		if parentErr != nil {
			// The parent does not exist either, so nothing can be created here.
			return "", ErrNotFound
		}
		if !fs.contains(resolvedParent) {
			return "", ErrAccessDenied
		}
		return filepath.Join(resolvedParent, filepath.Base(candidate)), nil
	}

	if !fs.contains(resolved) {
		return "", ErrAccessDenied
	}
	return resolved, nil
}

// hostPathNoFollow maps a resolved share-relative path to a host path without
// resolving a symbolic link in its final element.
//
// The operations that act on a name rather than on contents -- deleting and
// renaming -- must act on the name the client gave. Resolving its final element
// makes them act on whatever it points at instead: a delete unlinks the target
// and leaves the link, and a rename moves the target out from under it.
//
// Containment is unaffected. Every component but the last is still resolved and
// the result checked against the root, which is what hostPath relies on for a
// target that does not exist yet, and a link can only be traversed through a
// component that exists. The final element is not traversed here at all -- it is
// the thing being operated on -- so it cannot lead anywhere.
func (fs *LocalFileSystem) hostPathNoFollow(path string) (string, error) {
	if path == "" {
		return fs.root, nil
	}

	candidate := filepath.Join(fs.root, filepath.FromSlash(path))

	resolvedParent, err := filepath.EvalSymlinks(filepath.Dir(candidate))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNotFound
		}
		return "", err
	}
	if !fs.contains(resolvedParent) {
		return "", ErrAccessDenied
	}
	return filepath.Join(resolvedParent, filepath.Base(candidate)), nil
}

// contains reports whether a resolved host path is the root or lies beneath it.
//
// The separator on the prefix matters: without it, a sibling directory whose name
// merely starts with the root's name would compare as inside it.
func (fs *LocalFileSystem) contains(resolved string) bool {
	if resolved == fs.root {
		return true
	}
	return strings.HasPrefix(resolved, fs.root+string(os.PathSeparator))
}

// translate maps a host error onto the sentinel the server understands.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		return ErrNotFound
	case errors.Is(err, os.ErrExist):
		return ErrExists
	case errors.Is(err, os.ErrPermission):
		return ErrAccessDenied
	}
	// A few conditions have no portable errors.Is target.
	text := err.Error()
	switch {
	case strings.Contains(text, "directory not empty"):
		return ErrNotEmpty
	case strings.Contains(text, "not a directory"):
		return ErrNotDirectory
	case strings.Contains(text, "is a directory"):
		return ErrIsDirectory
	}
	return err
}

// attrOf renders a host entry.
func attrOf(name string, info os.FileInfo) FileAttr {
	modified := info.ModTime().UTC()
	return FileAttr{
		Name:  name,
		IsDir: info.IsDir(),
		Size:  info.Size(),
		// The host does not report an allocation size portably, so it is the
		// length rounded up to a 512-byte boundary, which is what a client uses
		// it for.
		AllocationSize: (info.Size() + 511) / 512 * 512,
		ReadOnly:       info.Mode().Perm()&0o200 == 0,
		// Only one timestamp is portable, so it stands for all four rather than
		// reporting a zero a client would render as the year 1601.
		Created:  modified,
		Accessed: modified,
		Modified: modified,
		Changed:  modified,
	}
}

// Open opens or creates a file.
func (fs *LocalFileSystem) Open(path string, flags OpenFlags) (File, error) {
	if fs.readOnly && (flags.Write || flags.Create || flags.CreateNew || flags.Truncate) {
		return nil, ErrReadOnly
	}

	host, err := fs.hostPath(path)
	if err != nil {
		return nil, err
	}

	// A directory is opened for metadata only: the host refuses to open one for
	// reading on some platforms, and SMB only ever uses such a handle to query.
	if info, statErr := os.Stat(host); statErr == nil && info.IsDir() {
		if flags.NonDirectory {
			return nil, ErrIsDirectory
		}
		if flags.CreateNew {
			return nil, ErrExists
		}
		return &localFile{path: host, directory: true}, nil
	} else if statErr == nil && flags.Directory {
		return nil, ErrNotDirectory
	}

	mode := os.O_RDONLY
	switch {
	case flags.Read && flags.Write:
		mode = os.O_RDWR
	case flags.Write:
		mode = os.O_RDWR
	}
	if flags.CreateNew {
		mode |= os.O_CREATE | os.O_EXCL
	} else if flags.Create {
		mode |= os.O_CREATE
	}
	if flags.Truncate {
		mode |= os.O_TRUNC
	}

	if flags.Directory {
		// Reaching here means the target does not exist: an existing directory
		// returned a handle above, and an existing file was refused there. So this
		// branch creates one, and may only be taken when the disposition asked for
		// a create.
		//
		// Without that condition an open-existing-only request created the
		// directory it was meant to find, and did so on a read-only share as well:
		// the read-only guards above test Write, Create, CreateNew and Truncate,
		// none of which a bare FILE_OPEN with FILE_DIRECTORY_FILE sets.
		if !flags.Create {
			return nil, ErrNotFound
		}
		if err := os.Mkdir(host, 0o750); err != nil {
			return nil, translate(err)
		}
		return &localFile{path: host, directory: true}, nil
	}

	file, err := os.OpenFile(host, mode, 0o640)
	if err != nil {
		return nil, translate(err)
	}
	return &localFile{file: file, path: host}, nil
}

// Stat describes a path.
func (fs *LocalFileSystem) Stat(path string) (FileAttr, error) {
	host, err := fs.hostPath(path)
	if err != nil {
		return FileAttr{}, err
	}
	info, err := os.Stat(host)
	if err != nil {
		return FileAttr{}, translate(err)
	}
	_, name := splitPath(path)
	if name == "" {
		name = fs.label
	}
	return attrOf(name, info), nil
}

// SetAttr applies the selected fields.
func (fs *LocalFileSystem) SetAttr(path string, attr FileAttr, mask AttrMask) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	host, err := fs.hostPath(path)
	if err != nil {
		return err
	}

	if mask.Size {
		if err := os.Truncate(host, attr.Size); err != nil {
			return translate(err)
		}
	}
	if mask.ReadOnly {
		info, err := os.Stat(host)
		if err != nil {
			return translate(err)
		}
		perm := info.Mode().Perm()
		if attr.ReadOnly {
			perm &= ^os.FileMode(0o222)
		} else {
			perm |= 0o200
		}
		if err := os.Chmod(host, perm); err != nil {
			return translate(err)
		}
	}
	// The host carries only an access and a modification time, so the two SMB
	// timestamps that map onto them are applied and the others ignored rather
	// than faked.
	if mask.Accessed || mask.Modified {
		accessed, modified := attr.Accessed, attr.Modified
		if !mask.Accessed || accessed.IsZero() {
			accessed = time.Now()
		}
		if !mask.Modified || modified.IsZero() {
			modified = time.Now()
		}
		if err := os.Chtimes(host, accessed, modified); err != nil {
			return translate(err)
		}
	}
	return nil
}

// Remove deletes a file.
//
// The name is resolved without following a link in its final element, so
// deleting a symbolic link removes the link rather than what it points at.
func (fs *LocalFileSystem) Remove(path string) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	host, err := fs.hostPathNoFollow(path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(host)
	if err != nil {
		return translate(err)
	}
	if info.IsDir() {
		return ErrIsDirectory
	}
	return translate(os.Remove(host))
}

// Rename moves a file or directory.
//
// Neither name has a link in its final element followed: renaming a symbolic link
// moves the link, and renaming onto one replaces the link rather than writing
// through it.
func (fs *LocalFileSystem) Rename(oldPath, newPath string, replace bool) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	oldHost, err := fs.hostPathNoFollow(oldPath)
	if err != nil {
		return err
	}
	newHost, err := fs.hostPathNoFollow(newPath)
	if err != nil {
		return err
	}
	if !replace {
		if _, err := os.Lstat(newHost); err == nil {
			return ErrExists
		}
	}
	return translate(os.Rename(oldHost, newHost))
}

// Mkdir creates a directory.
func (fs *LocalFileSystem) Mkdir(path string) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	host, err := fs.hostPath(path)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(host); err == nil {
		return ErrExists
	}
	return translate(os.Mkdir(host, 0o750))
}

// Rmdir removes an empty directory.
func (fs *LocalFileSystem) Rmdir(path string) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	if path == "" {
		return ErrAccessDenied
	}
	host, err := fs.hostPathNoFollow(path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(host)
	if err != nil {
		return translate(err)
	}
	// A symbolic link is not a directory, whatever it points at, so removing a
	// directory through one is refused rather than removing the directory it
	// names. Remove is what unlinks the link itself.
	if !info.IsDir() {
		return ErrNotDirectory
	}
	return translate(os.Remove(host))
}

// ReadDir lists the entries of a directory that match pattern.
func (fs *LocalFileSystem) ReadDir(path, pattern string) ([]DirEntry, error) {
	host, err := fs.hostPath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(host)
	if err != nil {
		return nil, translate(err)
	}
	if !info.IsDir() {
		return nil, ErrNotDirectory
	}

	held, err := os.ReadDir(host)
	if err != nil {
		return nil, translate(err)
	}

	listed := []DirEntry{}
	for _, entry := range held {
		if !matchPattern(pattern, entry.Name()) {
			continue
		}
		// A name the resolver would refuse is not listed either: a client that
		// saw it could not then open it, and it may be a link out of the share.
		if _, err := resolvePath(joinPath(path, entry.Name())); err != nil {
			continue
		}
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		listed = append(listed, DirEntry{Attr: attrOf(entry.Name(), entryInfo)})
	}

	sort.Slice(listed, func(i, j int) bool { return listed[i].Attr.Name < listed[j].Attr.Name })
	return listed, nil
}

// VolumeInfo describes the storage.
func (fs *LocalFileSystem) VolumeInfo() (VolumeInfo, error) {
	return VolumeInfo{
		Label:          fs.label,
		FileSystemName: "NTFS",
		// Derived from the root so it is stable across restarts, which is what a
		// client caches it for.
		SerialNumber:             serialNumberOf(fs.root),
		SectorsPerAllocationUnit: 8,
		BytesPerSector:           512,
	}, nil
}

// serialNumberOf derives a stable volume serial from a path.
func serialNumberOf(path string) uint32 {
	// FNV-1a, inline to avoid pulling in a hash for four lines.
	const offset, prime = uint32(2166136261), uint32(16777619)
	sum := offset
	for i := 0; i < len(path); i++ {
		sum ^= uint32(path[i])
		sum *= prime
	}
	return sum
}

// localFile is an open handle onto a host file. A directory handle carries no
// file, because SMB only ever uses one to query metadata.
type localFile struct {
	file      *os.File
	path      string
	directory bool
}

func (f *localFile) ReadAt(p []byte, off int64) (int, error) {
	if f.directory {
		return 0, ErrIsDirectory
	}
	read, err := f.file.ReadAt(p, off)
	// A short read at the end of the file is not a failure: SMB reports the short
	// count, so only a read that got nothing is an end-of-file.
	if err == io.EOF && read > 0 {
		return read, nil
	}
	if err != nil && err != io.EOF {
		return read, translate(err)
	}
	return read, err
}

func (f *localFile) WriteAt(p []byte, off int64) (int, error) {
	if f.directory {
		return 0, ErrIsDirectory
	}
	written, err := f.file.WriteAt(p, off)
	return written, translate(err)
}

func (f *localFile) Truncate(size int64) error {
	if f.directory {
		return ErrIsDirectory
	}
	return translate(f.file.Truncate(size))
}

func (f *localFile) Sync() error {
	if f.directory {
		return nil
	}
	return translate(f.file.Sync())
}

func (f *localFile) Stat() (FileAttr, error) {
	info, err := os.Stat(f.path)
	if err != nil {
		return FileAttr{}, translate(err)
	}
	return attrOf(filepath.Base(f.path), info), nil
}

func (f *localFile) Close() error {
	if f.file == nil {
		return nil
	}
	return translate(f.file.Close())
}

// Compile-time assurance that the backend satisfies the contracts.
var (
	_ FileSystem = (*LocalFileSystem)(nil)
	_ File       = (*localFile)(nil)
)
