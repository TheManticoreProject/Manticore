package server

import (
	"encoding/binary"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// preNTSearchMagic marks a resume key as one this server issued.
//
// A client hands back a key the server gave it, and the server has to tell its own
// key from a value a client invented. Without the marker an arbitrary 21 bytes
// would be read as a position and the search would resume somewhere meaningless.
var preNTSearchMagic = [4]byte{'M', 'A', 'N', 'T'}

// preNTSearchLimit bounds how many entries one response may carry, whatever the
// client asked for. Each entry is a fixed 43 bytes, so this is also a bound on the
// response size.
const preNTSearchLimit = 512

// handleSearch answers SMB_COM_SEARCH: the core-set directory enumeration.
//
// The search keeps no state on the server. The position travels in the resume key,
// which the client returns with its next request, and the directory is read again
// each time. That is what the command is shaped for — it has no close, so a search
// holding server state would leak one allocation per client that walked away
// mid-listing — and it is what a real server does with it.
//
// The cost is that the listing is a fresh view each call rather than a snapshot,
// so an entry created or removed between calls may be seen twice or missed. The
// command offers no way to avoid that; TRANS2_FIND_FIRST2, which does keep a
// snapshot, is the level a client uses when it matters.
//
// Wire format: [MS-CIFS] section 2.2.4.58.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleSearch(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.SearchRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	entries, next, status := conn.preNTSearch(req,
		decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode()),
		int(request.MaxCount),
		request.ResumeKeyPresent, request.ResumeKey)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// The response's data block is BufferFormat(0x05) DataLength(2) followed by
	// the packed entries, which is exactly what an SMB_STRING in variable-block
	// form emits.
	packed := []byte{}
	for index := range entries {
		entries[index].ResumeKey = preNTResumeKey(next, request.ResumeKey)
		encoded, err := entries[index].Marshal()
		if err != nil {
			logger.Debugf("SMB1 server: failed to encode a search entry for %s: %v", conn.Remote, err)
			return nt_status.NT_STATUS_UNSUCCESSFUL
		}
		packed = append(packed, encoded...)
	}

	response := commands.NewSearchResponse()
	response.Count = types.USHORT(len(entries))
	response.SMB_Directory_Information.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK)
	response.SMB_Directory_Information.Buffer = []types.UCHAR(packed)

	// An empty listing is STATUS_NO_MORE_FILES rather than an empty success:
	// that is how the command says the walk is finished, and a client given an
	// empty successful answer asks again forever.
	if len(entries) == 0 {
		return nt_status.NT_STATUS_NO_MORE_FILES
	}

	return conn.answer(w, response)
}

// handleFind answers SMB_COM_FIND, which carries the same request and response as
// SMB_COM_SEARCH and differs only in being closable with SMB_COM_FIND_CLOSE.
//
// Since the search here holds no server state, the close has nothing to release —
// see handleFindClose.
//
// Wire format: [MS-CIFS] section 2.2.4.59.
func handleFind(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.FindRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	resumeKey, present := preNTResumeKeyFrom(request.ResumeKey.Buffer)

	entries, next, status := conn.preNTSearch(req,
		decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode()),
		int(request.MaxCount), present, resumeKey)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if len(entries) == 0 {
		return nt_status.NT_STATUS_NO_MORE_FILES
	}

	for index := range entries {
		entries[index].ResumeKey = preNTResumeKey(next, resumeKey)
	}

	response := commands.NewFindResponse()
	response.Count = types.USHORT(len(entries))
	response.DirectoryInformationData = entries
	return conn.answer(w, response)
}

// handleFindUnique answers SMB_COM_FIND_UNIQUE: one listing with no continuation.
//
// [MS-CIFS] section 2.2.4.60 — the command exists to look up a single name, so no
// resume key is accepted and none is meaningful in the answer.
func handleFindUnique(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.FindUniqueRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	entries, _, status := conn.preNTSearch(req,
		decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode()),
		int(request.MaxCount), false, types.SMB_RESUME_KEY{})
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if len(entries) == 0 {
		return nt_status.NT_STATUS_NO_SUCH_FILE
	}

	response := commands.NewFindUniqueResponse()
	response.Count = types.USHORT(len(entries))
	response.DirectoryInformationData = entries
	return conn.answer(w, response)
}

// handleFindClose answers SMB_COM_FIND_CLOSE.
//
// There is nothing to release: a pre-NT search carries its position in the resume
// key rather than in a server-side table, so a client that stops asking has
// already freed everything the search used. Answering success is accurate rather
// than a stub — a client closing a search it can no longer continue is exactly
// what the command is for, and reporting an error would make an orderly client
// look like it had made a mistake.
//
// Wire format: [MS-CIFS] section 2.2.4.61.
func handleFindClose(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	if _, ok := req.Command.(*commands.FindCloseRequest); !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	// The TID still has to name a tree: closing a search on a share the client
	// does not hold is a protocol error whether or not anything is released.
	if _, status := conn.treeFor(req); status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	return conn.answer(w, commands.NewFindCloseResponse())
}

// preNTSearch enumerates a directory for the core-set search commands.
//
// Parameters:
//   - req: the request, for its TID and encoding
//   - pattern: the search path and wildcard the client sent
//   - maxCount: how many entries the client will take
//   - resuming: whether a resume key accompanied the request
//   - resumeKey: that key
//
// Returns:
//   - The entries, the position to report in their resume keys, and a status
func (c *Connection) preNTSearch(
	req *message.Message,
	pattern string,
	maxCount int,
	resuming bool,
	resumeKey types.SMB_RESUME_KEY,
) ([]types.SMB_DIRECTORY_INFORMATION, int, nt_status.NT_STATUS) {
	tree, status := c.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, 0, status
	}
	if tree.Share.FS == nil {
		return nil, 0, nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	directory, wildcard, err := resolvePathPattern(pattern)
	if err != nil {
		logger.Debugf("SMB1 server: %s searched %q, which is refused: %v", c.Remote, pattern, err)
		return nil, 0, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// A path with no wildcard names one entry rather than a directory to walk.
	// SMB_COM_FIND_UNIQUE exists precisely to look up a single name, and the
	// other two accept one as well, so reading the named path as a directory
	// would answer STATUS_NOT_A_DIRECTORY for the command's main use.
	var found []DirEntry
	if wildcard == "" {
		attr, err := tree.Share.FS.Stat(directory)
		if err != nil {
			return nil, 0, statusForFSError(err)
		}
		// Stat describes the path, and the entry has to name it: the listing
		// carries base names, not the path the client asked about.
		attr.Name = baseNameOf(directory)
		found = []DirEntry{{Attr: attr}}
	} else {
		found, err = tree.Share.FS.ReadDir(directory, wildcard)
		if err != nil {
			return nil, 0, statusForFSError(err)
		}
	}

	// Where to start. A key this server did not issue is refused rather than
	// guessed at: resuming from a position derived from arbitrary bytes would
	// return an arbitrary part of the directory and call it a continuation.
	start := 0
	if resuming {
		position, ours := preNTPositionOf(resumeKey)
		if !ours {
			logger.Debugf("SMB1 server: %s resumed a search with a key this server did not issue", c.Remote)
			return nil, 0, nt_status.NT_STATUS_INVALID_PARAMETER
		}
		start = position
	}
	if start > len(found) {
		start = len(found)
	}

	if maxCount <= 0 || maxCount > preNTSearchLimit {
		maxCount = preNTSearchLimit
	}

	entries := []types.SMB_DIRECTORY_INFORMATION{}
	position := start
	skipped := 0

	for position < len(found) && len(entries) < maxCount {
		attr := found[position].Attr
		position++

		// The entry's name field is a fixed 13 bytes: twelve for an 8.3 name and
		// one for its terminator. A longer name cannot be carried, and this
		// server has no 8.3 alias to substitute — inventing one would hand the
		// client a name that no subsequent open could resolve, which is worse
		// than the file not appearing. Such entries are skipped, and a client
		// that needs to see them has to use TRANS2_FIND_FIRST2.
		if !fitsShortNameField(attr.Name) {
			skipped++
			continue
		}

		entry := types.NewSMB_DIRECTORY_INFORMATION()
		entry.FileAttributes = types.UCHAR(legacyAttributesFor(attr))
		entry.LastWriteDate, entry.LastWriteTime = dosDateTimeOf(attr.Modified)
		entry.FileSize = types.ULONG(uint32(attr.Size))
		entry.FileName = *types.NewOEM_STRINGFromString(attr.Name)
		entries = append(entries, *entry)
	}

	if skipped > 0 {
		logger.Debugf("SMB1 server: omitted %d entries of %q from a core-set search for %s, "+
			"their names not fitting the 8.3 field", skipped, directory, c.Remote)
	}

	return entries, position, nt_status.NT_STATUS_SUCCESS
}

// fitsShortNameField reports whether a name can be carried in an entry's
// twelve-byte 8.3 name field.
//
// The check is on the encoded length rather than on the shape of the name: a
// name of eight characters and a three-character extension is the intent, but
// anything that fits and round-trips is serviceable, and anything that does not
// fit cannot be sent whatever its shape.
func fitsShortNameField(name string) bool {
	if len(name) == 0 || len(name) > 12 {
		return false
	}
	// A name with more than one dot, or a dot in the wrong place, still fits the
	// field and still opens, so it is allowed through. Only a name carrying a
	// character the field cannot hold is refused.
	return !strings.ContainsRune(name, 0x00)
}

// preNTResumeKey builds the resume key an entry reports, carrying the position to
// continue from.
//
// ClientState is echoed from the request: [MS-CIFS] section 2.2.4.58.1 says "The
// value provided by the client MUST be returned in each ResumeKey provided in the
// response", so it is the client's to keep and not the server's to use.
func preNTResumeKey(position int, from types.SMB_RESUME_KEY) types.SMB_RESUME_KEY {
	key := types.SMB_RESUME_KEY{ClientState: from.ClientState}
	binary.LittleEndian.PutUint32(key.ServerState[0:4], uint32(position))
	copy(key.ServerState[4:8], preNTSearchMagic[:])
	return key
}

// preNTPositionOf reads the position out of a resume key, and reports whether the
// key is one this server issued.
func preNTPositionOf(key types.SMB_RESUME_KEY) (int, bool) {
	if !equalBytes(key.ServerState[4:8], preNTSearchMagic[:]) {
		return 0, false
	}
	return int(binary.LittleEndian.Uint32(key.ServerState[0:4])), true
}

// preNTResumeKeyFrom decodes the resume key SMB_COM_FIND carries in its data
// block, which is modelled as a string rather than as the fixed structure.
func preNTResumeKeyFrom(raw []types.UCHAR) (types.SMB_RESUME_KEY, bool) {
	key := types.SMB_RESUME_KEY{}
	if len(raw) < types.SMB_RESUME_KEY_SIZE {
		return key, false
	}
	if _, err := key.Unmarshal([]byte(raw)); err != nil {
		return types.SMB_RESUME_KEY{}, false
	}
	return key, true
}

// equalBytes compares two byte slices, avoiding a dependency for one comparison.
func equalBytes(a []types.UCHAR, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if byte(a[index]) != b[index] {
			return false
		}
	}
	return true
}

// baseNameOf returns the final component of a share-relative path, which is the
// name a listing entry carries.
func baseNameOf(path string) string {
	if index := strings.LastIndexAny(path, `/\`); index >= 0 {
		return path[index+1:]
	}
	return path
}
