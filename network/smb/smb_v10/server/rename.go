package server

import (
	"strings"
	"sync"
)

// openRegistry holds every handle open on one share, whichever connection it
// belongs to.
//
// A connection's own table of handles belongs to its receive loop and cannot be
// read from anywhere else, and a rename has to reach the handles of every
// connection: a handle that kept the old name would apply its delete-on-close,
// and every later set by handle, to whatever file is created under that name
// afterwards, and leave the renamed file alone.
type openRegistry struct {
	mutex sync.Mutex
	opens map[*Open]struct{}
}

// newOpenRegistry builds an empty registry.
func newOpenRegistry() *openRegistry {
	return &openRegistry{opens: make(map[*Open]struct{})}
}

// add records a handle. A nil registry, which a share built outside AddShare
// has, records nothing.
func (r *openRegistry) add(open *Open) {
	if r == nil {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.opens[open] = struct{}{}
}

// remove forgets a handle.
func (r *openRegistry) remove(open *Open) {
	if r == nil {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.opens, open)
}

// rename renames a file or directory on the share and moves every handle on it
// to the new name. Every rename command comes through here.
//
// A rename that would replace a file some handle still holds is refused with
// ErrAccessDenied, which answers STATUS_ACCESS_DENIED. [MS-FSA] section
// 2.1.5.15.11 fails a replacing rename with STATUS_ACCESS_DENIED when
// "TargetLink.File.OpenList contains an Open" and there is no oplock to break
// first, and [MS-FSCC] section 2.4.42 lists STATUS_ACCESS_DENIED for "the
// target file was open and ReplaceIfExists is nonzero". The level II oplocks
// this server grants never make a holder close, so the target is refused
// whether or not one is held. Replacing it anyway would leave its handles
// naming the renamed file: their delete-on-close would delete it, and their
// locks would be merged with its own.
//
// Parameters:
//   - from: the share-relative path to rename
//   - to: the share-relative path to rename it to
//   - replace: whether an existing file at to is replaced
//
// Returns:
//   - The backend's error, or ErrAccessDenied for an open target
func (s *Share) rename(from, to string, replace bool) error {
	if replace && from != to && s.opens.holdsAt(to) {
		return ErrAccessDenied
	}
	if err := s.FS.Rename(from, to, replace); err != nil {
		return err
	}
	s.renameOpens(from, to)
	return nil
}

// holdsAt reports whether any handle names a path or lies beneath it.
func (r *openRegistry) holdsAt(path string) bool {
	if r == nil {
		return false
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for open := range r.opens {
		if _, under := renamedPath(open.currentPath(), path, path); under {
			return true
		}
	}
	return false
}

// renameOpens moves every handle on a renamed file to its new name: the file
// itself and, when it is a directory, everything beneath it.
//
// The lock and oplock entries those handles own move in the same step. Both
// tables file a handle under the path it reports, and they are held throughout,
// so no table operation can see a handle under one name and its entry under the
// other.
//
// Parameters:
//   - from: the share-relative path that was renamed
//   - to: the share-relative path it was renamed to
func (s *Share) renameOpens(from, to string) {
	if s == nil || s.opens == nil {
		return
	}
	if s.oplocks != nil {
		s.oplocks.mutex.Lock()
		defer s.oplocks.mutex.Unlock()
	}
	if s.locks != nil {
		s.locks.mutex.Lock()
		defer s.locks.mutex.Unlock()
	}
	s.opens.mutex.Lock()
	defer s.opens.mutex.Unlock()

	for open := range s.opens.opens {
		old := open.currentPath()
		moved, ok := renamedPath(old, from, to)
		if !ok {
			continue
		}
		if s.oplocks != nil {
			s.oplocks.moveLocked(open, old, moved)
		}
		if s.locks != nil {
			s.locks.moveLocked(open, old, moved)
		}
		open.renamed.Store(&moved)
	}
}

// renamedPath returns where a path is after a rename, and whether the rename
// moved it at all: the path is the one renamed, or lies beneath it.
//
// Parameters:
//   - path: the share-relative path a handle reports
//   - from: the share-relative path that was renamed
//   - to: the share-relative path it was renamed to
//
// The match is exact, case included, because that is how the backends resolve a
// rename's source: the in-memory one keys entries by the exact path, and the
// local one hands the path to the host file system unchanged. A handle opened
// under a different spelling of the name than the rename used is therefore not
// moved.
//
// Returns:
//   - The path's new value, and whether the rename changed it
func renamedPath(path, from, to string) (string, bool) {
	if path == from {
		return to, true
	}
	if from != "" && strings.HasPrefix(path, from+"/") {
		return to + path[len(from):], true
	}
	return "", false
}

// moveLocked refiles the oplock a handle holds from one path to another. The
// caller holds the mutex.
func (t *oplockTable) moveLocked(open *Open, from, to string) {
	fromKey, toKey := oplockKey(from), oplockKey(to)
	if fromKey == toKey {
		// A rename that only changes case leaves the key where it was.
		return
	}
	held := t.granted[fromKey]
	kept := held[:0]
	for _, holder := range held {
		if holder.owner == open {
			t.granted[toKey] = append(t.granted[toKey], holder)
			continue
		}
		kept = append(kept, holder)
	}
	if len(kept) == 0 {
		delete(t.granted, fromKey)
	} else {
		t.granted[fromKey] = kept
	}
}

// moveLocked refiles the byte-range locks a handle owns from one path to
// another. The caller holds the mutex.
func (t *lockTable) moveLocked(open *Open, from, to string) {
	if from == to {
		return
	}
	held := t.granted[from]
	kept := held[:0]
	for _, lock := range held {
		if lock.owner == open {
			t.granted[to] = append(t.granted[to], lock)
			continue
		}
		kept = append(kept, lock)
	}
	if len(kept) == 0 {
		delete(t.granted, from)
	} else {
		t.granted[from] = kept
	}
}
