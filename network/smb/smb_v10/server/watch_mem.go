package server

import (
	"strings"
	"sync"

	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// memoryWatch is one registered watch on a MemoryFileSystem.
type memoryWatch struct {
	// path is the directory being watched, share-relative, and recursive says
	// whether entries below it count.
	path      string
	recursive bool

	// changes carries the notifications. It is buffered, and a send that would
	// block is dropped: a watch that fell behind must not hold up the operation
	// that made the change.
	changes chan ChangeNotification
}

// memoryWatchBuffer is how many notifications a watch holds before it starts
// dropping them.
//
// A client that misses one re-enumerates, which is what the protocol tells it to
// do when more changed than fits in a response, so a bounded buffer loses nothing
// a client cannot recover.
const memoryWatchBuffer = 64

// watchRegistry holds the watches on a file system and publishes to them.
//
// It is a separate type so that both backends share the fan-out and the dropping
// rule rather than each implementing it.
type watchRegistry struct {
	mutex   sync.Mutex
	watches map[*memoryWatch]struct{}
}

// add registers a watch and returns it with the function that removes it.
func (r *watchRegistry) add(path string, recursive bool) (<-chan ChangeNotification, func(), error) {
	watch := &memoryWatch{
		path:      strings.Trim(path, "/"),
		recursive: recursive,
		changes:   make(chan ChangeNotification, memoryWatchBuffer),
	}

	r.mutex.Lock()
	if r.watches == nil {
		r.watches = make(map[*memoryWatch]struct{})
	}
	r.watches[watch] = struct{}{}
	r.mutex.Unlock()

	var once sync.Once
	stop := func() {
		once.Do(func() {
			r.mutex.Lock()
			delete(r.watches, watch)
			r.mutex.Unlock()
			close(watch.changes)
		})
	}
	return watch.changes, stop, nil
}

// publish tells every interested watch that a path changed.
//
// Parameters:
//   - path: the share-relative path that changed
//   - action: the FILE_ACTION_* value describing what happened
func (r *watchRegistry) publish(path string, action uint32) {
	r.mutex.Lock()
	interested := make([]*memoryWatch, 0, len(r.watches))
	for watch := range r.watches {
		if watch.covers(path) {
			interested = append(interested, watch)
		}
	}
	r.mutex.Unlock()

	for _, watch := range interested {
		notification := ChangeNotification{
			Action: action,
			Name:   watch.relative(path),
		}
		// Non-blocking: a full buffer drops the notification rather than
		// stalling the operation that made the change.
		select {
		case watch.changes <- notification:
		default:
		}
	}
}

// covers reports whether a changed path is one this watch reports.
//
// A non-recursive watch reports only its immediate children, which is what
// WatchTree being clear means: a change two levels down is not a change to the
// directory the client asked about.
func (w *memoryWatch) covers(path string) bool {
	relative := w.relative(path)
	if relative == "" {
		return false
	}
	if w.recursive {
		return true
	}
	return !strings.Contains(relative, "/")
}

// relative renders a changed path relative to the watched directory, in the
// backslash-separated form FILE_NOTIFY_INFORMATION carries.
func (w *memoryWatch) relative(path string) string {
	trimmed := strings.Trim(path, "/")
	if w.path == "" {
		return strings.ReplaceAll(trimmed, "/", `\`)
	}
	if !strings.HasPrefix(trimmed, w.path+"/") {
		return ""
	}
	return strings.ReplaceAll(strings.TrimPrefix(trimmed, w.path+"/"), "/", `\`)
}

// Watch begins watching a directory of the memory file system.
//
// The directory has to exist and be a directory: watching something that is not
// there would report changes to a path the client cannot enumerate.
func (fs *MemoryFileSystem) Watch(path string, recursive bool) (<-chan ChangeNotification, func(), error) {
	if path != "" {
		attr, err := fs.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		if !attr.IsDir {
			return nil, nil, ErrNotDirectory
		}
	}
	return fs.watches.add(path, recursive)
}

// The FILE_ACTION_* values a change carries, named from windows/filesystem so the
// two agree.
var (
	changeAdded    = uint32(filesystem.FileActionAdded)
	changeRemoved  = uint32(filesystem.FileActionRemoved)
	changeModified = uint32(filesystem.FileActionModified)
)

// count reports how many watches are registered, for the tests and for a caller
// that wants to report on a share.
func (r *watchRegistry) count() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return len(r.watches)
}
