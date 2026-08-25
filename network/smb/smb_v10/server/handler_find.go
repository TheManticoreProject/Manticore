package server

import (
	"encoding/binary"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// Search flags a client sends on a FIND ([MS-CIFS] 2.2.6.2.1).
const (
	// findClose closes the search once this response is sent.
	findClose = 0x0001
	// findCloseIfEndOfSearch closes it when the enumeration finishes.
	findCloseIfEndOfSearch = 0x0002
	// findContinueFromLast resumes from the server's own cursor rather than from
	// a resume key the client supplies.
	findContinueFromLast = 0x0008
)

// Search is a directory enumeration in progress.
//
// The whole listing is taken once, when the search opens, and then handed out in
// batches. Re-reading the directory on each continuation would be worse than it
// sounds: entries appearing or vanishing between batches would make the cursor
// mean something different each time, so a client could see an entry twice or miss
// one entirely. A snapshot is stale but coherent, which is the trade a client
// enumerating a directory expects.
type Search struct {
	// SID is the identifier the client sends to continue the search.
	SID uint16

	// Tree is the tree the search runs on, and Directory the resolved directory
	// being listed.
	Tree      *Tree
	Directory string

	// Pattern is what names are matched against.
	Pattern string

	// InformationLevel is the shape the entries are returned in, fixed when the
	// search opens: a continuation that asked for a different one would be
	// describing a different enumeration.
	InformationLevel uint16

	// Entries is the snapshot, and Position how far through it the client has
	// been taken.
	Entries  []DirEntry
	Position int

	Created time.Time
}

// exhausted reports whether every entry has been handed out.
func (s *Search) exhausted() bool {
	return s.Position >= len(s.Entries)
}

// Search returns the enumeration a SID names, or nil when it names none.
func (c *Connection) Search(sid uint16) *Search {
	return c.searches[sid]
}

// closeSearch drops an enumeration and releases its identifier.
func (c *Connection) closeSearch(sid uint16) bool {
	if _, ok := c.searches[sid]; !ok {
		return false
	}
	delete(c.searches, sid)
	c.sids.Release(sid)
	return true
}

// handleFindFirst2 answers TRANS2_FIND_FIRST2: it opens an enumeration and returns
// its first batch.
//
// Request parameters: SearchAttributes(2) SearchCount(2) Flags(2)
// InformationLevel(2) SearchStorageType(4) FileName(variable).
func handleFindFirst2(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 12 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	searchCount := int(binary.LittleEndian.Uint16(parameters[2:4]))
	flags := binary.LittleEndian.Uint16(parameters[4:6])
	informationLevel := binary.LittleEndian.Uint16(parameters[6:8])
	// The pattern is OEM or UTF-16 according to the request's own flag, like every
	// other name a client sends. Reading it as OEM regardless turns a Unicode
	// "\\*" into a backslash, a null and a star, which the path resolver refuses.
	requested := decodeWireString([]types.UCHAR(parameters[12:]), req.Header.Flags2.IsUnicode())

	if !supportedFindLevel(informationLevel) {
		logger.Debugf("SMB1 server: %s asked for find information level 0x%04X, which is not served",
			conn.Remote, informationLevel)
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	directory, pattern, err := resolvePathPattern(requested)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked to enumerate %q, which is refused: %v", conn.Remote, requested, err)
		return nil, nil, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// A search whose pattern holds no wildcard resolves to a single path, and that
	// path may name a file rather than a directory: naming one file exactly is how
	// a client asks about it without opening it, and a client does that before
	// deleting or renaming something. So a non-directory is turned back into a
	// literal pattern matched in its parent, rather than being listed as a
	// directory it is not.
	if pattern == "" && directory != "" {
		if attr, statErr := tree.Share.FS.Stat(directory); statErr == nil && !attr.IsDir {
			directory, pattern = splitFinalElement(directory)
		}
	}

	entries, err := tree.Share.FS.ReadDir(directory, pattern)
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	// "." and ".." are what a client expects at the head of a listing, and some
	// clients rely on them to recognise a directory. They are synthesised rather
	// than expected from a backend, since a backend has no reason to invent them.
	entries = append(dotEntries(tree, directory), entries...)

	sid, err := conn.sids.Allocate()
	if err != nil {
		logger.Warnf("SMB1 server: refusing a search from %s: %v", conn.Remote, err)
		return nil, nil, nt_status.NT_STATUS_OS2_NO_MORE_SIDS
	}

	search := &Search{
		SID:              sid,
		Tree:             tree,
		Directory:        directory,
		Pattern:          pattern,
		InformationLevel: informationLevel,
		Entries:          entries,
		Created:          time.Now().UTC(),
	}

	data, returned := encodeFindEntries(search, searchCount, reassembly.maxDataCount, req.Header.Flags2.IsUnicode())
	endOfSearch := search.exhausted()

	// The search is kept only while there is more to hand out, and only if the
	// client did not ask for it to be closed.
	keep := !endOfSearch && flags&findClose == 0
	if endOfSearch && flags&findCloseIfEndOfSearch != 0 {
		keep = false
	}
	if keep {
		conn.searches[sid] = search
	} else {
		conn.sids.Release(sid)
	}

	logger.Debugf("SMB1 server: %s enumerated %q on %q: %d entries, end=%t, sid=0x%04X",
		conn.Remote, directory, tree.Share.Name, returned, endOfSearch, sid)

	// Response parameters: SID(2) SearchCount(2) EndOfSearch(2) EaErrorOffset(2)
	// LastNameOffset(2).
	responseParameters := make([]byte, 10)
	binary.LittleEndian.PutUint16(responseParameters[0:2], sid)
	binary.LittleEndian.PutUint16(responseParameters[2:4], uint16(returned))
	binary.LittleEndian.PutUint16(responseParameters[4:6], boolToUint16(endOfSearch))

	return responseParameters, data, nt_status.NT_STATUS_SUCCESS
}

// handleFindNext2 answers TRANS2_FIND_NEXT2: it hands out the next batch of an
// enumeration already open.
//
// Request parameters: SID(2) SearchCount(2) InformationLevel(2) ResumeKey(4)
// Flags(2) FileName(variable).
func handleFindNext2(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 12 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	sid := binary.LittleEndian.Uint16(parameters[0:2])
	searchCount := int(binary.LittleEndian.Uint16(parameters[2:4]))
	informationLevel := binary.LittleEndian.Uint16(parameters[4:6])
	flags := binary.LittleEndian.Uint16(parameters[10:12])

	search := conn.Search(sid)
	if search == nil {
		logger.Debugf("SMB1 server: %s continued search 0x%04X, which is not open", conn.Remote, sid)
		return nil, nil, nt_status.NT_STATUS_INVALID_HANDLE
	}
	// The search belongs to the tree it was opened on.
	if search.Tree.TID != uint16(req.Header.TID) {
		return nil, nil, nt_status.NT_STATUS_INVALID_HANDLE
	}
	// A continuation that changed the shape would be describing a different
	// enumeration, so it is refused rather than quietly honoured.
	if informationLevel != search.InformationLevel {
		logger.Debugf("SMB1 server: %s continued search 0x%04X at level 0x%04X, opened at 0x%04X",
			conn.Remote, sid, informationLevel, search.InformationLevel)
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	data, returned := encodeFindEntries(search, searchCount, reassembly.maxDataCount, req.Header.Flags2.IsUnicode())
	endOfSearch := search.exhausted()

	if flags&findClose != 0 || (endOfSearch && flags&findCloseIfEndOfSearch != 0) {
		conn.closeSearch(sid)
	}

	// Response parameters: SearchCount(2) EndOfSearch(2) EaErrorOffset(2)
	// LastNameOffset(2). No SID: the client already has it.
	responseParameters := make([]byte, 8)
	binary.LittleEndian.PutUint16(responseParameters[0:2], uint16(returned))
	binary.LittleEndian.PutUint16(responseParameters[2:4], boolToUint16(endOfSearch))

	return responseParameters, data, nt_status.NT_STATUS_SUCCESS
}

// handleFindClose2 answers SMB_COM_FIND_CLOSE2: it releases a search handle.
func handleFindClose2(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.FindClose2Request)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	if !conn.closeSearch(uint16(request.SearchHandle)) {
		return nt_status.NT_STATUS_INVALID_HANDLE
	}

	return conn.answer(w, commands.NewFindClose2Response())
}

// dotEntries synthesises the "." and ".." entries at the head of a listing. ".."
// is omitted at the share root, where there is no parent within the share to
// name.
func dotEntries(tree *Tree, directory string) []DirEntry {
	self, err := tree.Share.FS.Stat(directory)
	if err != nil {
		// A directory that cannot be described still lists; the timestamps are
		// the only thing lost.
		self = FileAttr{IsDir: true}
	}

	self.Name = "."
	entries := []DirEntry{{Attr: self}}

	if directory == "" {
		return entries
	}

	parent, _ := splitPath(directory)
	up, err := tree.Share.FS.Stat(parent)
	if err != nil {
		up = FileAttr{IsDir: true}
	}
	up.Name = ".."
	return append(entries, DirEntry{Attr: up})
}

// boolToUint16 renders a flag as the 16-bit value the wire carries.
func boolToUint16(value bool) uint16 {
	if value {
		return 1
	}
	return 0
}
