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

	// rootDir is the share directory opened as an os.Root. Opening, creating,
	// stating, listing and removing go through it: each path component is
	// opened relative to the one before without following links blindly, and a
	// link is followed only when its target stays beneath the root. A directory
	// swapped for a link out of the share between two calls therefore cannot
	// redirect the operation. A link with an absolute target is refused even
	// when that target is inside the share, as os.Root requires.
	rootDir *os.Root

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

	rootDir, err := os.OpenRoot(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", resolved, err)
	}

	return &LocalFileSystem{root: resolved, rootDir: rootDir, label: label}, nil
}

// rel turns a resolved share-relative path into the name the os.Root takes:
// host separators, and "." for the share root itself.
func rel(path string) string {
	if path == "" {
		return "."
	}
	return filepath.FromSlash(path)
}

// SetReadOnly refuses every modifying operation at the backend.
func (fs *LocalFileSystem) SetReadOnly(readOnly bool) {
	fs.readOnly = readOnly
}

// Close releases the handle held on the share directory. The file system is
// unusable afterwards.
func (fs *LocalFileSystem) Close() error {
	return fs.rootDir.Close()
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
// is checked instead — which is what a create needs. That is only sufficient when
// the final element is truly absent: a dangling symbolic link also fails to
// resolve, yet a create through it makes the file at the link's target, wherever
// that is. Such a link is refused, since its target cannot be checked against the
// root until it exists, and by then the create has already escaped.
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
		// root.
		parent := filepath.Dir(candidate)
		resolvedParent, parentErr := filepath.EvalSymlinks(parent)
		if parentErr != nil {
			// The parent does not exist either, so nothing can be created here.
			return "", ErrNotFound
		}
		if !fs.contains(resolvedParent) {
			return "", ErrAccessDenied
		}
		host := filepath.Join(resolvedParent, filepath.Base(candidate))
		// The final element must be absent rather than a dangling link: opening
		// one with O_CREATE creates its target, and a link planted in the share,
		// or a relative one a client has renamed to a shallower directory, points
		// out of it.
		if _, lstatErr := os.Lstat(host); !errors.Is(lstatErr, os.ErrNotExist) {
			return "", ErrAccessDenied
		}
		return host, nil
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
	case strings.Contains(text, "too many levels of symbolic links"):
		// A link loop, or O_NOFOLLOW refusing a link in the final element.
		return ErrAccessDenied
	case strings.Contains(text, "no such device or address"):
		// A FIFO opened for writing with nothing reading it.
		return ErrAccessDenied
	case strings.Contains(text, "path escapes from parent"):
		// os.Root refused a path that resolves outside the share, or a link
		// with an absolute target.
		return ErrAccessDenied
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

	name := rel(path)

	// A directory is opened for metadata only: the host refuses to open one for
	// reading on some platforms, and SMB only ever uses such a handle to query.
	info, statErr := fs.rootDir.Stat(name)
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return nil, translate(statErr)
	}
	if statErr == nil && info.IsDir() {
		if flags.NonDirectory {
			return nil, ErrIsDirectory
		}
		if flags.CreateNew {
			return nil, ErrExists
		}
		return &localFile{fs: fs, name: name, directory: true}, nil
	} else if statErr == nil && flags.Directory {
		return nil, ErrNotDirectory
	} else if statErr == nil && !info.Mode().IsRegular() {
		// A FIFO, socket or device is not a file a client can use, and opening
		// a FIFO would block until its other end is opened.
		return nil, ErrAccessDenied
	} else if statErr == nil && info.Mode().Perm()&0o200 == 0 && (flags.Write || flags.Truncate) {
		// A file carrying the read-only attribute refuses an open for writing
		// and an overwrite with STATUS_ACCESS_DENIED, [MS-FSA] section
		// 2.1.5.1.2.1. The host enforces that only for an unprivileged
		// process: one running as root truncated the file below.
		return nil, ErrAccessDenied
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
		if err := fs.rootDir.Mkdir(name, 0o750); err != nil {
			return nil, translate(err)
		}
		return &localFile{fs: fs, name: name, directory: true}, nil
	}

	// A create through a dangling link would make the file at the link's
	// target. os.Root keeps that target inside the share, but the name the
	// client asked for would still not be the file created, so it is refused.
	if errors.Is(statErr, os.ErrNotExist) && (flags.Create || flags.CreateNew) {
		if linkInfo, err := fs.rootDir.Lstat(name); err == nil && linkInfo.Mode()&os.ModeSymlink != 0 {
			return nil, ErrAccessDenied
		}
	}

	// The open resolves every component beneath the root, so neither a link in
	// the path nor a directory swapped for one between requests can lead out of
	// the share.
	file, err := fs.openRegular(name, mode, 0o640, false)
	if err != nil {
		return nil, err
	}
	return &localFile{fs: fs, file: file, name: name}, nil
}

// errNotRegular refuses a FIFO, socket or device, which no client operation can
// use.
var errNotRegular = fmt.Errorf("%w: not a regular file", ErrAccessDenied)

// openRegular opens name beneath the root and refuses anything but a regular
// file, or a directory when allowDir is set. The open does not block on a FIFO,
// and the type is checked on the handle, so an entry replaced by a FIFO after an
// earlier stat is still refused.
func (fs *LocalFileSystem) openRegular(name string, mode int, perm os.FileMode, allowDir bool) (*os.File, error) {
	file, err := fs.rootDir.OpenFile(name, mode|openNonBlock, perm)
	if err != nil {
		return nil, translate(err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, translate(err)
	}
	if !info.Mode().IsRegular() && !(allowDir && info.IsDir()) {
		file.Close()
		if info.IsDir() {
			return nil, ErrIsDirectory
		}
		return nil, errNotRegular
	}
	return file, nil
}

// Link gives an existing file a second name.
//
// Both paths go through hostPathNoFollow, so neither the source nor the target
// can be placed outside the share by way of a symbolic link — the same rule the
// name operations follow, and for the same reason: a link is created by name, and
// a name that resolves through a link out of the share would put a file where the
// share does not reach.
//
// Parameters:
//   - existing: the file to link to, share-relative
//   - target: the new name, share-relative
//
// Returns:
//   - An error if either path is unusable or the host refuses the link
func (fs *LocalFileSystem) Link(existing, target string) error {
	if fs.readOnly {
		return ErrReadOnly
	}

	source, err := fs.hostPathNoFollow(existing)
	if err != nil {
		return err
	}
	destination, err := fs.hostPathNoFollow(target)
	if err != nil {
		return err
	}

	info, err := os.Lstat(source)
	if err != nil {
		return translate(err)
	}
	if info.IsDir() {
		return ErrIsDirectory
	}

	if err := os.Link(source, destination); err != nil {
		return translate(err)
	}
	return nil
}

// Stat describes a path.
func (fs *LocalFileSystem) Stat(path string) (FileAttr, error) {
	info, err := fs.rootDir.Stat(rel(path))
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
	name := rel(path)

	// The size and read-only attribute are changed through a handle opened
	// beneath the root, so the file changed is the one the share names.
	if mask.Size {
		file, err := fs.openRegular(name, os.O_WRONLY, 0, false)
		if err != nil {
			return err
		}
		err = file.Truncate(attr.Size)
		file.Close()
		if err != nil {
			return translate(err)
		}
	}
	if mask.ReadOnly {
		if err := fs.setReadOnly(path, attr.ReadOnly); err != nil {
			return err
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
		// os.Root has no Chtimes before Go 1.25, so this one step is still
		// path-based (#1896).
		host, err := fs.hostPath(path)
		if err != nil {
			return err
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
	name := rel(path)
	info, err := fs.rootDir.Lstat(name)
	if err != nil {
		return translate(err)
	}
	if info.IsDir() {
		return ErrIsDirectory
	}
	// Unlinking a host file needs write access to its directory, not to the file,
	// so the read-only attribute Stat reports has to be honoured here.
	if info.Mode().Perm()&0o200 == 0 {
		return ErrCannotDelete
	}
	return translate(fs.rootDir.Remove(name))
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
	// The host refuses a directory moved beneath itself with an error of its
	// own, which says nothing a client can act on. Refused here, it reads the
	// same as from any other backend.
	if isBeneath(newPath, oldPath) {
		if info, err := os.Lstat(oldHost); err == nil && info.IsDir() {
			return ErrAccessDenied
		}
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
	name := rel(path)
	if _, err := fs.rootDir.Lstat(name); err == nil {
		return ErrExists
	}
	return translate(fs.rootDir.Mkdir(name, 0o750))
}

// Rmdir removes an empty directory.
func (fs *LocalFileSystem) Rmdir(path string) error {
	if fs.readOnly {
		return ErrReadOnly
	}
	if path == "" {
		return ErrAccessDenied
	}
	name := rel(path)
	info, err := fs.rootDir.Lstat(name)
	if err != nil {
		return translate(err)
	}
	// A symbolic link is not a directory, whatever it points at, so removing a
	// directory through one is refused rather than removing the directory it
	// names. Remove is what unlinks the link itself.
	if !info.IsDir() {
		return ErrNotDirectory
	}
	return translate(fs.rootDir.Remove(name))
}

// ReadDir lists the entries of a directory that match pattern.
func (fs *LocalFileSystem) ReadDir(path, pattern string) ([]DirEntry, error) {
	dir, err := fs.rootDir.OpenFile(rel(path), os.O_RDONLY|openNonBlock, 0)
	if err != nil {
		return nil, translate(err)
	}
	defer dir.Close()
	info, err := dir.Stat()
	if err != nil {
		return nil, translate(err)
	}
	if !info.IsDir() {
		return nil, ErrNotDirectory
	}

	held, err := dir.ReadDir(-1)
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
	fs        *LocalFileSystem
	file      *os.File
	name      string // root-relative name
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
	var info os.FileInfo
	var err error
	if f.file != nil {
		info, err = f.file.Stat()
	} else {
		info, err = f.fs.rootDir.Stat(f.name)
	}
	if err != nil {
		return FileAttr{}, translate(err)
	}
	name := filepath.Base(f.name)
	if f.name == "." {
		// The share root reports the directory's own name, as before.
		name = filepath.Base(f.fs.root)
	}
	return attrOf(name, info), nil
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
