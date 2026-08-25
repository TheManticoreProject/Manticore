package server

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ShareType is the Service string a tree connect reports, describing what kind of
// resource the tree names ([MS-CIFS] 2.2.4.55.2).
type ShareType string

const (
	// ShareTypeDisk is a file-system share.
	ShareTypeDisk ShareType = "A:"
	// ShareTypeNamedPipe is the IPC$ share, over which named pipes are reached.
	ShareTypeNamedPipe ShareType = "IPC"
	// ShareTypePrinter is a print queue.
	ShareTypePrinter ShareType = "LPT1:"
	// ShareTypeAny matches any type, which a client may send to mean "whatever
	// this name is".
	ShareTypeAny ShareType = "?????"
)

// Share is a named resource a client can connect a tree to.
type Share struct {
	// Name is the share name, matched case-insensitively as Windows does.
	Name string

	// Type is what the share is. A disk share requires FS to be set.
	Type ShareType

	// Comment is the description a share enumeration would report.
	Comment string

	// ReadOnly refuses every modifying operation on the share, whatever access
	// the client asked for. It is enforced in the handlers rather than left to
	// the backend, so a backend cannot forget it.
	ReadOnly bool

	// FS is the storage behind a disk share.
	FS FileSystem
}

// OpenFlags describe what an open is for. They are the subset of the client's
// requested access and options that a backend needs in order to open the file:
// share modes and the finer access bits are enforced above the backend.
type OpenFlags struct {
	// Read and Write are the access the open needs.
	Read  bool
	Write bool

	// Create allows the open to create the file if it does not exist, and
	// CreateNew requires that it did not exist.
	Create    bool
	CreateNew bool

	// Truncate empties an existing file on open.
	Truncate bool

	// Directory requires the target to be a directory, and NonDirectory requires
	// that it is not. Both set is a contradiction and is refused above.
	Directory    bool
	NonDirectory bool
}

// FileAttr describes a file or directory.
type FileAttr struct {
	// Name is the entry's own name, without any path.
	Name string

	// IsDir reports whether the entry is a directory.
	IsDir bool

	// Size is the length in bytes, and AllocationSize the space reserved for it.
	// A backend that does not distinguish the two reports the same value twice.
	Size           int64
	AllocationSize int64

	// ReadOnly reports that the entry cannot be modified.
	ReadOnly bool

	// Created, Accessed, Modified and Changed are the four timestamps SMB
	// carries. A backend without a distinct value for one repeats another.
	Created  time.Time
	Accessed time.Time
	Modified time.Time
	Changed  time.Time
}

// AttrMask selects which fields of a FileAttr a SetAttr call applies, so that
// setting one timestamp does not overwrite the others with zeroes.
type AttrMask struct {
	Size     bool
	ReadOnly bool
	Created  bool
	Accessed bool
	Modified bool
	Changed  bool
}

// DirEntry is one entry of a directory listing.
type DirEntry struct {
	Attr FileAttr
}

// VolumeInfo describes the storage behind a share.
type VolumeInfo struct {
	// Label and FileSystemName are reported to a client that asks about the
	// volume.
	Label          string
	FileSystemName string

	// SerialNumber identifies the volume.
	SerialNumber uint32

	// TotalBytes and FreeBytes describe capacity. A backend that does not know
	// reports zero for both.
	TotalBytes int64
	FreeBytes  int64

	// SectorsPerAllocationUnit and BytesPerSector describe the allocation
	// geometry a client uses to turn sizes into cluster counts.
	SectorsPerAllocationUnit uint32
	BytesPerSector           uint32
}

// File is an open file on a FileSystem.
//
// Offsets are explicit on every call: SMB carries the offset in the request, so a
// backend needs no cursor of its own and two opens of the same file cannot
// interfere through a shared position.
type File interface {
	// ReadAt reads into p starting at off. It returns io.EOF only when nothing
	// could be read; a short read at the end of the file is not an error, because
	// SMB reports a short read rather than a failure.
	ReadAt(p []byte, off int64) (int, error)

	// WriteAt writes p at off, extending the file if needed.
	WriteAt(p []byte, off int64) (int, error)

	// Truncate sets the file's length.
	Truncate(size int64) error

	// Sync commits any buffered contents.
	Sync() error

	// Stat describes the open file.
	Stat() (FileAttr, error)

	// Close releases the handle.
	Close() error
}

// FileSystem is the storage behind a disk share.
//
// Every path reaching a FileSystem has already been resolved by the server:
// forward-slash separated, relative to the share root, with no empty, "." or ".."
// element, and never absolute. A backend does not have to defend against
// traversal, and MUST NOT try to interpret a path as anything other than that
// form — the containment guarantee lives in one place so it can be reviewed in
// one place.
//
// A backend reports failure with the sentinel errors below where they apply, so
// the server can map an outcome to the right protocol status.
type FileSystem interface {
	// Open opens or creates a file according to flags.
	Open(path string, flags OpenFlags) (File, error)

	// Stat describes a path without opening it.
	Stat(path string) (FileAttr, error)

	// SetAttr applies the fields of attr that mask selects.
	SetAttr(path string, attr FileAttr, mask AttrMask) error

	// Remove deletes a file. It refuses a directory, which Rmdir handles.
	Remove(path string) error

	// Rename moves a file or directory. When replace is false it fails if the
	// destination exists.
	Rename(oldPath, newPath string, replace bool) error

	// Mkdir creates a directory. Its parent must exist.
	Mkdir(path string) error

	// Rmdir removes an empty directory.
	Rmdir(path string) error

	// ReadDir lists the entries of a directory whose names match pattern, which
	// may contain the SMB wildcards "*" and "?". An empty pattern matches
	// everything.
	ReadDir(path, pattern string) ([]DirEntry, error)

	// VolumeInfo describes the storage.
	VolumeInfo() (VolumeInfo, error)
}

// Sentinel errors a FileSystem returns so the server can answer with the right
// protocol status. A backend may wrap them.
var (
	// ErrNotFound reports that the path does not exist.
	ErrNotFound = fmt.Errorf("no such file or directory")

	// ErrExists reports that the path already exists where it must not.
	ErrExists = fmt.Errorf("file or directory already exists")

	// ErrNotDirectory reports that a path element that had to be a directory is
	// not one, or that a directory operation was asked of a file.
	ErrNotDirectory = fmt.Errorf("not a directory")

	// ErrIsDirectory reports that a file operation was asked of a directory.
	ErrIsDirectory = fmt.Errorf("is a directory")

	// ErrNotEmpty reports that a directory being removed still has entries.
	ErrNotEmpty = fmt.Errorf("directory not empty")

	// ErrAccessDenied reports that the operation is not permitted.
	ErrAccessDenied = fmt.Errorf("access denied")

	// ErrReadOnly reports that the share or the entry refuses modification.
	ErrReadOnly = fmt.Errorf("read-only")
)

// AddShare registers a share on the server. Share names are matched
// case-insensitively, so two that differ only in case collide.
//
// Parameters:
//   - share: the share to serve
//
// Returns:
//   - An error if the share is not usable or its name is already taken
func (s *Server) AddShare(share *Share) error {
	if share == nil {
		return fmt.Errorf("cannot add a nil share")
	}
	if share.Name == "" {
		return fmt.Errorf("a share must have a name")
	}
	if strings.ContainsAny(share.Name, `\/:*?"<>|`) {
		return fmt.Errorf("share name %q contains a character a share name cannot carry", share.Name)
	}
	if share.Type == "" {
		share.Type = ShareTypeDisk
	}
	if share.Type == ShareTypeDisk && share.FS == nil {
		return fmt.Errorf("disk share %q has no file system", share.Name)
	}

	key := strings.ToUpper(share.Name)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.shares == nil {
		s.shares = make(map[string]*Share)
	}
	if _, taken := s.shares[key]; taken {
		return fmt.Errorf("a share named %q is already registered", share.Name)
	}
	s.shares[key] = share
	return nil
}

// Share returns the share a name refers to, or nil when none does.
func (s *Server) Share(name string) *Share {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.shares[strings.ToUpper(name)]
}

// Shares returns the registered shares.
func (s *Server) Shares() []*Share {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registered := make([]*Share, 0, len(s.shares))
	for _, share := range s.shares {
		registered = append(registered, share)
	}
	return registered
}

// matchPattern reports whether a name matches an SMB wildcard pattern, where "*"
// stands for any run of characters and "?" for exactly one. Matching is
// case-insensitive, as the file system commands are.
//
// An empty pattern, and the "*" and "*.*" forms a client sends to mean
// "everything", match every name.
func matchPattern(pattern, name string) bool {
	if pattern == "" || pattern == "*" || pattern == "*.*" {
		return true
	}
	return wildcardMatch(strings.ToUpper(pattern), strings.ToUpper(name))
}

// wildcardMatch matches a pattern against a name, both already folded to one
// case. It walks the two with a backtrack point for the most recent "*", which
// matches without recursion and cannot be made to blow up by a pattern full of
// stars.
func wildcardMatch(pattern, name string) bool {
	patternIndex, nameIndex := 0, 0
	starPattern, starName := -1, 0

	for nameIndex < len(name) {
		switch {
		case patternIndex < len(pattern) && (pattern[patternIndex] == '?' || pattern[patternIndex] == name[nameIndex]):
			patternIndex++
			nameIndex++
		case patternIndex < len(pattern) && pattern[patternIndex] == '*':
			starPattern = patternIndex
			starName = nameIndex
			patternIndex++
		case starPattern >= 0:
			// Backtrack: let the last star swallow one more character.
			patternIndex = starPattern + 1
			starName++
			nameIndex = starName
		default:
			return false
		}
	}

	// Any remaining pattern must be stars.
	for patternIndex < len(pattern) && pattern[patternIndex] == '*' {
		patternIndex++
	}
	return patternIndex == len(pattern)
}

// errEOF is returned by a backend read that reached the end of the file. It is
// separate from the sentinels above because it is not a failure.
var errEOF = io.EOF
