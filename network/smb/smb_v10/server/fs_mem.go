package server

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryFileSystem is a FileSystem held entirely in memory.
//
// It exists for two reasons. Tests get a backend with no disk and no cleanup, so
// a file-service test is about the protocol rather than about a temporary
// directory. And a caller can serve a share that touches no storage at all —
// useful when the point is to look like a file server rather than to be one.
//
// It is safe for concurrent use.
type MemoryFileSystem struct {
	mutex sync.RWMutex

	// entries holds every file and directory by its resolved path. The share root
	// is the empty path and is always present.
	entries map[string]*memoryEntry

	// volume describes the storage reported to a client.
	volume VolumeInfo
}

// memoryEntry is one file or directory.
type memoryEntry struct {
	name  string
	isDir bool
	data  []byte

	readOnly bool
	created  time.Time
	accessed time.Time
	modified time.Time
	changed  time.Time
}

// NewMemoryFileSystem creates an empty in-memory file system.
//
// Parameters:
//   - label: the volume label reported to a client
//
// Returns:
//   - The file system, containing only its root
func NewMemoryFileSystem(label string) *MemoryFileSystem {
	now := time.Now().UTC()
	return &MemoryFileSystem{
		entries: map[string]*memoryEntry{
			"": {name: "", isDir: true, created: now, accessed: now, modified: now, changed: now},
		},
		volume: VolumeInfo{
			Label:                    label,
			FileSystemName:           "NTFS",
			SerialNumber:             0x4D454D30,
			TotalBytes:               1 << 30,
			FreeBytes:                1 << 30,
			SectorsPerAllocationUnit: 8,
			BytesPerSector:           512,
		},
	}
}

// AddFile seeds a file, creating the directories above it. It is for a caller
// setting up a share before serving it, and for tests.
//
// Parameters:
//   - path: the share-relative path, slash separated
//   - contents: the file's contents
//
// Returns:
//   - An error if the path is unusable
func (fs *MemoryFileSystem) AddFile(path string, contents []byte) error {
	resolved, err := resolvePath(path)
	if err != nil {
		return err
	}
	if resolved == "" {
		return fmt.Errorf("cannot add a file at the share root")
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err := fs.ensureParents(resolved); err != nil {
		return err
	}

	now := time.Now().UTC()
	_, name := splitPath(resolved)
	stored := make([]byte, len(contents))
	copy(stored, contents)
	fs.entries[resolved] = &memoryEntry{
		name: name, data: stored,
		created: now, accessed: now, modified: now, changed: now,
	}
	return nil
}

// AddDirectory seeds a directory, creating the directories above it.
func (fs *MemoryFileSystem) AddDirectory(path string) error {
	resolved, err := resolvePath(path)
	if err != nil {
		return err
	}
	if resolved == "" {
		return nil
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err := fs.ensureParents(resolved); err != nil {
		return err
	}
	return fs.makeDirectory(resolved)
}

// ensureParents creates every missing directory above a path. The caller holds
// the lock.
func (fs *MemoryFileSystem) ensureParents(path string) error {
	parent, _ := splitPath(path)
	if parent == "" {
		return nil
	}

	built := ""
	for _, element := range strings.Split(parent, "/") {
		built = joinPath(built, element)
		if existing, ok := fs.entries[built]; ok {
			if !existing.isDir {
				return fmt.Errorf("%q is a file, so it cannot hold %q: %w", built, path, ErrNotDirectory)
			}
			continue
		}
		if err := fs.makeDirectory(built); err != nil {
			return err
		}
	}
	return nil
}

// makeDirectory records a directory. The caller holds the lock.
func (fs *MemoryFileSystem) makeDirectory(path string) error {
	now := time.Now().UTC()
	_, name := splitPath(path)
	fs.entries[path] = &memoryEntry{
		name: name, isDir: true,
		created: now, accessed: now, modified: now, changed: now,
	}
	return nil
}

// attrOf renders an entry. The caller holds the lock.
func (e *memoryEntry) attr() FileAttr {
	return FileAttr{
		Name:           e.name,
		IsDir:          e.isDir,
		Size:           int64(len(e.data)),
		AllocationSize: int64(len(e.data)),
		ReadOnly:       e.readOnly,
		Created:        e.created,
		Accessed:       e.accessed,
		Modified:       e.modified,
		Changed:        e.changed,
	}
}

// Open opens or creates a file.
func (fs *MemoryFileSystem) Open(path string, flags OpenFlags) (File, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	entry, exists := fs.entries[path]

	switch {
	case exists && flags.CreateNew:
		return nil, ErrExists
	case !exists && !flags.Create:
		return nil, ErrNotFound
	}

	if exists {
		if entry.isDir && flags.NonDirectory {
			return nil, ErrIsDirectory
		}
		if !entry.isDir && flags.Directory {
			return nil, ErrNotDirectory
		}
		if flags.Write && entry.readOnly {
			return nil, ErrReadOnly
		}
		if flags.Truncate && !entry.isDir {
			entry.data = nil
			entry.modified = time.Now().UTC()
		}
		return &memoryFile{fs: fs, path: path}, nil
	}

	// Creating. The parent has to exist: a create does not build a tree.
	parent, name := splitPath(path)
	if parent != "" {
		holder, ok := fs.entries[parent]
		if !ok {
			return nil, ErrNotFound
		}
		if !holder.isDir {
			return nil, ErrNotDirectory
		}
	}

	now := time.Now().UTC()
	fs.entries[path] = &memoryEntry{
		name: name, isDir: flags.Directory,
		created: now, accessed: now, modified: now, changed: now,
	}
	return &memoryFile{fs: fs, path: path}, nil
}

// Stat describes a path.
func (fs *MemoryFileSystem) Stat(path string) (FileAttr, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	entry, ok := fs.entries[path]
	if !ok {
		return FileAttr{}, ErrNotFound
	}
	return entry.attr(), nil
}

// SetAttr applies the selected fields.
func (fs *MemoryFileSystem) SetAttr(path string, attr FileAttr, mask AttrMask) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	entry, ok := fs.entries[path]
	if !ok {
		return ErrNotFound
	}

	if mask.Size && !entry.isDir {
		fs.resize(entry, attr.Size)
	}
	if mask.ReadOnly {
		entry.readOnly = attr.ReadOnly
	}
	if mask.Created {
		entry.created = attr.Created
	}
	if mask.Accessed {
		entry.accessed = attr.Accessed
	}
	if mask.Modified {
		entry.modified = attr.Modified
	}
	if mask.Changed {
		entry.changed = attr.Changed
	}
	return nil
}

// resize grows or shrinks an entry's contents. The caller holds the lock.
func (fs *MemoryFileSystem) resize(entry *memoryEntry, size int64) {
	switch {
	case size < 0:
		return
	case size <= int64(len(entry.data)):
		entry.data = entry.data[:size]
	default:
		grown := make([]byte, size)
		copy(grown, entry.data)
		entry.data = grown
	}
	entry.modified = time.Now().UTC()
}

// Remove deletes a file.
func (fs *MemoryFileSystem) Remove(path string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	entry, ok := fs.entries[path]
	if !ok {
		return ErrNotFound
	}
	if entry.isDir {
		return ErrIsDirectory
	}
	if entry.readOnly {
		return ErrReadOnly
	}
	delete(fs.entries, path)
	return nil
}

// Rename moves a file or directory, and everything beneath a directory with it.
func (fs *MemoryFileSystem) Rename(oldPath, newPath string, replace bool) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	entry, ok := fs.entries[oldPath]
	if !ok {
		return ErrNotFound
	}
	if _, taken := fs.entries[newPath]; taken {
		if !replace {
			return ErrExists
		}
		delete(fs.entries, newPath)
	}

	// The destination's parent has to exist, or the entry would be unreachable.
	if parent, _ := splitPath(newPath); parent != "" {
		holder, ok := fs.entries[parent]
		if !ok {
			return ErrNotFound
		}
		if !holder.isDir {
			return ErrNotDirectory
		}
	}

	// Move the entry, then everything beneath it if it is a directory. Collected
	// first, because the map is being written while it is walked.
	_, newName := splitPath(newPath)
	entry.name = newName
	fs.entries[newPath] = entry
	delete(fs.entries, oldPath)

	if !entry.isDir {
		return nil
	}
	prefix := oldPath + "/"
	moved := map[string]*memoryEntry{}
	for path, held := range fs.entries {
		if strings.HasPrefix(path, prefix) {
			moved[newPath+"/"+strings.TrimPrefix(path, prefix)] = held
		}
	}
	for path := range fs.entries {
		if strings.HasPrefix(path, prefix) {
			delete(fs.entries, path)
		}
	}
	for path, held := range moved {
		fs.entries[path] = held
	}
	return nil
}

// Mkdir creates a directory.
func (fs *MemoryFileSystem) Mkdir(path string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if _, taken := fs.entries[path]; taken {
		return ErrExists
	}
	if parent, _ := splitPath(path); parent != "" {
		holder, ok := fs.entries[parent]
		if !ok {
			return ErrNotFound
		}
		if !holder.isDir {
			return ErrNotDirectory
		}
	}
	return fs.makeDirectory(path)
}

// Rmdir removes an empty directory.
func (fs *MemoryFileSystem) Rmdir(path string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	entry, ok := fs.entries[path]
	if !ok {
		return ErrNotFound
	}
	if !entry.isDir {
		return ErrNotDirectory
	}
	if path == "" {
		return ErrAccessDenied
	}

	prefix := path + "/"
	for held := range fs.entries {
		if strings.HasPrefix(held, prefix) {
			return ErrNotEmpty
		}
	}
	delete(fs.entries, path)
	return nil
}

// ReadDir lists the immediate children of a directory that match pattern.
func (fs *MemoryFileSystem) ReadDir(path, pattern string) ([]DirEntry, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	holder, ok := fs.entries[path]
	if !ok {
		return nil, ErrNotFound
	}
	if !holder.isDir {
		return nil, ErrNotDirectory
	}

	prefix := ""
	if path != "" {
		prefix = path + "/"
	}

	listed := []DirEntry{}
	for held, entry := range fs.entries {
		if held == path || !strings.HasPrefix(held, prefix) {
			continue
		}
		// Immediate children only.
		if strings.Contains(strings.TrimPrefix(held, prefix), "/") {
			continue
		}
		if !matchPattern(pattern, entry.name) {
			continue
		}
		listed = append(listed, DirEntry{Attr: entry.attr()})
	}

	// A stable order, since a map has none and a client enumerating twice should
	// see the same thing.
	sort.Slice(listed, func(i, j int) bool { return listed[i].Attr.Name < listed[j].Attr.Name })
	return listed, nil
}

// VolumeInfo describes the storage.
func (fs *MemoryFileSystem) VolumeInfo() (VolumeInfo, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	return fs.volume, nil
}

// memoryFile is an open handle onto a MemoryFileSystem entry. It holds the path
// rather than the entry, so a rename or a delete underneath it is observed rather
// than written into a detached copy.
type memoryFile struct {
	fs     *MemoryFileSystem
	path   string
	closed bool
}

// entry resolves the handle's path, reporting the failure a vanished file gives.
func (f *memoryFile) entry() (*memoryEntry, error) {
	if f.closed {
		return nil, ErrAccessDenied
	}
	entry, ok := f.fs.entries[f.path]
	if !ok {
		return nil, ErrNotFound
	}
	return entry, nil
}

func (f *memoryFile) ReadAt(p []byte, off int64) (int, error) {
	f.fs.mutex.RLock()
	defer f.fs.mutex.RUnlock()

	entry, err := f.entry()
	if err != nil {
		return 0, err
	}
	if entry.isDir {
		return 0, ErrIsDirectory
	}
	if off < 0 {
		return 0, fmt.Errorf("negative offset %d", off)
	}
	if off >= int64(len(entry.data)) {
		return 0, io.EOF
	}
	return copy(p, entry.data[off:]), nil
}

func (f *memoryFile) WriteAt(p []byte, off int64) (int, error) {
	f.fs.mutex.Lock()
	defer f.fs.mutex.Unlock()

	entry, err := f.entry()
	if err != nil {
		return 0, err
	}
	if entry.isDir {
		return 0, ErrIsDirectory
	}
	if entry.readOnly {
		return 0, ErrReadOnly
	}
	if off < 0 {
		return 0, fmt.Errorf("negative offset %d", off)
	}

	// A write past the end extends the file, with the gap reading as zeroes.
	if end := off + int64(len(p)); end > int64(len(entry.data)) {
		grown := make([]byte, end)
		copy(grown, entry.data)
		entry.data = grown
	}
	written := copy(entry.data[off:], p)
	entry.modified = time.Now().UTC()
	return written, nil
}

func (f *memoryFile) Truncate(size int64) error {
	f.fs.mutex.Lock()
	defer f.fs.mutex.Unlock()

	entry, err := f.entry()
	if err != nil {
		return err
	}
	if entry.readOnly {
		return ErrReadOnly
	}
	f.fs.resize(entry, size)
	return nil
}

// Sync is a no-op: there is nothing behind memory to commit to.
func (f *memoryFile) Sync() error {
	if f.closed {
		return ErrAccessDenied
	}
	return nil
}

func (f *memoryFile) Stat() (FileAttr, error) {
	f.fs.mutex.RLock()
	defer f.fs.mutex.RUnlock()

	entry, err := f.entry()
	if err != nil {
		return FileAttr{}, err
	}
	return entry.attr(), nil
}

func (f *memoryFile) Close() error {
	f.closed = true
	return nil
}

// Compile-time assurance that the backend satisfies the contracts.
var (
	_ FileSystem = (*MemoryFileSystem)(nil)
	_ File       = (*memoryFile)(nil)
)
