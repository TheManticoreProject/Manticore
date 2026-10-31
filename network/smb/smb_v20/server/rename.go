package server

import (
	"strings"
	"sync"
)

// openRegistry records every open on the server, across connections, so a
// rename can move each open of the renamed file to its new name. Without it an
// open on another connection kept the old name, and its delete-on-close then
// removed whatever file was later created there.
type openRegistry struct {
	mutex sync.Mutex
	opens map[*Open]struct{}
}

func newOpenRegistry() *openRegistry {
	return &openRegistry{opens: make(map[*Open]struct{})}
}

// add records an open.
func (r *openRegistry) add(open *Open) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.opens[open] = struct{}{}
}

// remove forgets an open.
func (r *openRegistry) remove(open *Open) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.opens, open)
}

// holdsAt reports whether any open on the share names a path or lies beneath it.
func (r *openRegistry) holdsAt(share *Share, path string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for open := range r.opens {
		if _, under := renamedPath(open.currentPath(), path, path); open.tree.Share == share && under {
			return true
		}
	}
	return false
}

// renameOpens moves every open of a renamed file to its new name: the file
// itself and, for a directory, everything beneath it. The share-mode and oplock
// entries move in the same step, under their tables' mutexes, so the key each
// table files an open under always matches the path the open reports.
func (s *Server) renameOpens(share *Share, from, to string) {
	s.oplocks.mutex.Lock()
	defer s.oplocks.mutex.Unlock()
	s.shareModes.mutex.Lock()
	defer s.shareModes.mutex.Unlock()
	s.openFiles.mutex.Lock()
	defer s.openFiles.mutex.Unlock()

	for open := range s.openFiles.opens {
		if open.tree.Share != share {
			continue
		}
		old := open.currentPath()
		moved, ok := renamedPath(old, from, to)
		if !ok {
			continue
		}
		oldKey, newKey := fileKey(share, old), fileKey(share, moved)
		if _, held := s.shareModes.opens[oldKey][open]; held {
			delete(s.shareModes.opens[oldKey], open)
			if len(s.shareModes.opens[oldKey]) == 0 {
				delete(s.shareModes.opens, oldKey)
			}
			if s.shareModes.opens[newKey] == nil {
				s.shareModes.opens[newKey] = make(map[*Open]struct{})
			}
			s.shareModes.opens[newKey][open] = struct{}{}
		}
		if g := s.oplocks.grants[oldKey][open]; g != nil {
			delete(s.oplocks.grants[oldKey], open)
			if len(s.oplocks.grants[oldKey]) == 0 {
				delete(s.oplocks.grants, oldKey)
			}
			if s.oplocks.grants[newKey] == nil {
				s.oplocks.grants[newKey] = make(map[*Open]*oplockGrant)
			}
			s.oplocks.grants[newKey][open] = g
		}
		open.renamed.Store(&moved)
	}
}

// renamedPath returns where path is after from was renamed to to, and whether
// the rename moved it at all: path is from itself, or lies beneath it. The match
// is exact, case included, as the backends resolve a rename's source.
func renamedPath(path, from, to string) (string, bool) {
	if path == from {
		return to, true
	}
	if from != "" && strings.HasPrefix(path, from+"/") {
		return to + path[len(from):], true
	}
	return "", false
}
