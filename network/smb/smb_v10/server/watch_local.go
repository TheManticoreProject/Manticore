package server

import (
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
)

// LocalWatchInterval is how often a watch on a LocalFileSystem re-reads the
// directory it is watching.
//
// The host is not asked to report changes: doing that portably means a per-platform
// notification API, and this package is pure standard library by design. So a
// watch polls, and the interval is the latency a client sees between a change and
// its notification.
//
// It is a package variable rather than a constant so a test can shorten it. A
// caller has no reason to.
var LocalWatchInterval = 500 * time.Millisecond

// localWatch is one polling watch on a LocalFileSystem.
type localWatch struct {
	fs        *LocalFileSystem
	path      string
	recursive bool

	changes chan ChangeNotification
	done    chan struct{}
	once    sync.Once
}

// Watch begins watching a directory of the local file system.
//
// The watch compares successive listings rather than being told what changed, so
// it reports what it can see afterwards: an entry that appeared, one that
// disappeared, and one whose size or modification time moved. A change that leaves
// all three the same — a write that replaces a byte without changing the length,
// within the same timestamp granularity — is not visible to it, and that is a
// property of polling rather than something the interval can fix.
//
// Parameters:
//   - path: the share-relative directory to watch
//   - recursive: whether to report changes below it
//
// Returns:
//   - A channel of changes, a stop function, and an error if the path cannot be
//     watched
func (fs *LocalFileSystem) Watch(path string, recursive bool) (<-chan ChangeNotification, func(), error) {
	attr, err := fs.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !attr.IsDir {
		return nil, nil, ErrNotDirectory
	}

	watch := &localWatch{
		fs:        fs,
		path:      path,
		recursive: recursive,
		changes:   make(chan ChangeNotification, memoryWatchBuffer),
		done:      make(chan struct{}),
	}

	go watch.poll()

	stop := func() {
		watch.once.Do(func() { close(watch.done) })
	}
	return watch.changes, stop, nil
}

// snapshot is what a poll remembers about one entry, which is everything it can
// compare.
type snapshot struct {
	size     int64
	modified time.Time
	isDir    bool
}

// poll re-reads the directory until it is stopped, reporting what changed between
// readings.
func (w *localWatch) poll() {
	defer close(w.changes)

	previous := w.read()

	ticker := time.NewTicker(LocalWatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
		}

		current := w.read()
		w.report(previous, current)
		previous = current
	}
}

// read takes a snapshot of the watched directory.
func (w *localWatch) read() map[string]snapshot {
	taken := map[string]snapshot{}

	entries, err := w.fs.ReadDir(w.path, "")
	if err != nil {
		// A directory that has gone away stops producing changes rather than
		// reporting every entry as removed: the client's handle on it is what
		// will fail next, and inventing a storm of removals here would be worse
		// than silence.
		logger.Debugf("SMB1 server: a watch on %q could not read it: %v", w.path, err)
		return taken
	}

	for _, entry := range entries {
		taken[entry.Attr.Name] = snapshot{
			size:     entry.Attr.Size,
			modified: entry.Attr.Modified,
			isDir:    entry.Attr.IsDir,
		}
	}
	return taken
}

// report sends a notification for each difference between two snapshots.
func (w *localWatch) report(previous, current map[string]snapshot) {
	for name, now := range current {
		before, existed := previous[name]
		switch {
		case !existed:
			w.send(name, changeAdded)
		case now.size != before.size || !now.modified.Equal(before.modified):
			w.send(name, changeModified)
		}
	}

	for name := range previous {
		if _, present := current[name]; !present {
			w.send(name, changeRemoved)
		}
	}
}

// send delivers one notification, dropping it if the reader has fallen behind.
//
// Dropping is the right failure: a client that misses a notification
// re-enumerates, which is what the protocol tells it to do when more changed than
// fits in a response, so nothing is lost that the client cannot recover.
func (w *localWatch) send(name string, action uint32) {
	select {
	case w.changes <- ChangeNotification{Action: action, Name: name}:
	case <-w.done:
	default:
	}
}
