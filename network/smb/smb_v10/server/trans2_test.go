package server

import (
	"encoding/binary"
	"strings"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// namesOf renders a listing's names, so a test asserts on names rather than on
// whole entries.
func namesOf(entries []smb1client.Entry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.LongName)
	}
	return names
}

// TestDirectoryListing is the milestone this phase exists for: the SMB1 client in
// this repository lists a directory on the server.
func TestDirectoryListing(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	for _, name := range []string{"alpha.txt", "beta.txt", "gamma.log"} {
		if err := fs.AddFile(name, []byte(name)); err != nil {
			t.Fatalf("AddFile(%q) error = %v", name, err)
		}
	}
	if err := fs.AddDirectory("subdir"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	entries, err := client.ListDirectory("\\")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}

	names := namesOf(entries)
	for _, want := range []string{".", "..", "alpha.txt", "beta.txt", "gamma.log", "subdir"} {
		found := false
		for _, name := range names {
			if name == want {
				found = true
				break
			}
		}
		// ".." is absent at the share root, which has no parent within the share.
		if want == ".." {
			continue
		}
		if !found {
			t.Errorf("the listing is missing %q; it holds %v", want, names)
		}
	}

	// The directory is reported as one, and the files are not.
	for _, entry := range entries {
		switch entry.LongName {
		case "subdir", ".":
			if !entry.IsDirectory {
				t.Errorf("%q is not reported as a directory", entry.LongName)
			}
		case "alpha.txt":
			if entry.IsDirectory {
				t.Error("alpha.txt is reported as a directory")
			}
			if entry.Size != uint64(len("alpha.txt")) {
				t.Errorf("alpha.txt is reported as %d bytes, want %d", entry.Size, len("alpha.txt"))
			}
		}
	}
}

// TestDirectoryListingWithPattern asserts a wildcard filters the listing, and that
// a pattern matching nothing yields nothing rather than failing.
func TestDirectoryListingWithPattern(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	for _, name := range []string{"one.txt", "two.txt", "three.log"} {
		if err := fs.AddFile(name, []byte("x")); err != nil {
			t.Fatalf("AddFile(%q) error = %v", name, err)
		}
	}
	_, client := fileServer(t, fs, false)

	entries, err := client.ListEntries("\\*.txt")
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	names := namesOf(entries)

	// The dot entries are always present; the filter applies to the rest.
	for _, name := range names {
		if name == "." || name == ".." {
			continue
		}
		if !strings.HasSuffix(name, ".txt") {
			t.Errorf("the listing holds %q, which does not match *.txt", name)
		}
	}
	if !containsName(names, "one.txt") || !containsName(names, "two.txt") {
		t.Errorf("the listing is missing a match: %v", names)
	}
	if containsName(names, "three.log") {
		t.Errorf("the listing holds a non-match: %v", names)
	}

	// A pattern matching nothing still lists the dot entries and does not fail.
	entries, err = client.ListEntries("\\*.nothing")
	if err != nil {
		t.Fatalf("ListEntries() with a non-matching pattern error = %v", err)
	}
	for _, entry := range entries {
		if entry.LongName != "." && entry.LongName != ".." {
			t.Errorf("a non-matching pattern returned %q", entry.LongName)
		}
	}
}

// TestDirectoryListingInSubdirectory asserts a listing of something other than the
// share root works, and that ".." appears there.
func TestDirectoryListingInSubdirectory(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("dir/nested.txt", []byte("nested")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	entries, err := client.ListDirectory("\\dir")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
	names := namesOf(entries)

	if !containsName(names, "nested.txt") {
		t.Errorf("the listing is missing nested.txt: %v", names)
	}
	// A directory below the root has a parent to name.
	if !containsName(names, "..") {
		t.Errorf("the listing of a subdirectory has no %q entry: %v", "..", names)
	}
}

// TestDirectoryListingSpansSeveralBatches asserts an enumeration larger than one
// response is handed out across continuations, and that every entry arrives exactly
// once.
//
// This is what the search state exists for: the client asks again with the handle
// it was given, and the server picks up where it left off.
func TestDirectoryListingSpansSeveralBatches(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")

	// Enough entries, with long enough names, that the listing cannot fit in one
	// response.
	const count = 500
	expected := map[string]bool{}
	for i := 0; i < count; i++ {
		name := "a-file-with-a-reasonably-long-name-" + pad4(i) + ".txt"
		if err := fs.AddFile(name, []byte("x")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
		expected[name] = false
	}

	_, client := fileServer(t, fs, false)

	entries, err := client.ListDirectory("\\")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}

	// Prove the enumeration actually spanned batches rather than happening to fit.
	// Each entry is the 94-byte fixed part plus a name of about 45 bytes, so one
	// response's budget holds on the order of a hundred: a complete listing of 500
	// cannot have arrived in one.
	perBatch := (int(DefaultMaxBufferSize) - trans2ResponseOverhead) / (bothDirectoryInfoFixedSize + 48)
	if len(entries) <= perBatch {
		t.Fatalf("the listing returned %d entries, which could have fitted in one response of about %d; "+
			"the continuation path was not exercised", len(entries), perBatch)
	}

	for _, entry := range entries {
		if entry.LongName == "." || entry.LongName == ".." {
			continue
		}
		seen, known := expected[entry.LongName]
		if !known {
			t.Fatalf("the listing holds %q, which was never created", entry.LongName)
		}
		if seen {
			t.Fatalf("the listing holds %q twice", entry.LongName)
		}
		expected[entry.LongName] = true
	}

	missing := 0
	for name, seen := range expected {
		if !seen {
			if missing < 3 {
				t.Errorf("the listing is missing %q", name)
			}
			missing++
		}
	}
	if missing > 0 {
		t.Fatalf("%d of %d entries are missing from the listing", missing, count)
	}
}

// TestSearchHandleLifecycle asserts a search handle is released, and that a
// continuation or a close on a handle that is not open is refused.
func TestSearchHandleLifecycle(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	for i := 0; i < 200; i++ {
		if err := fs.AddFile("entry-"+pad4(i)+".txt", []byte("x")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
	}
	srv, client := fileServer(t, fs, false)

	// A full enumeration opens a handle and closes it on the way out.
	if _, err := client.ListDirectory("\\"); err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}

	waitFor(t, func() bool {
		for _, conn := range srv.liveConnections() {
			if len(conn.searches) != 0 {
				return false
			}
		}
		return true
	}, "the server still holds a search handle after the enumeration finished")

	// Several enumerations in a row do not leak handles either.
	for i := 0; i < 5; i++ {
		if _, err := client.ListDirectory("\\"); err != nil {
			t.Fatalf("ListDirectory() on pass %d error = %v", i, err)
		}
	}
	waitFor(t, func() bool {
		for _, conn := range srv.liveConnections() {
			if len(conn.searches) != 0 {
				return false
			}
		}
		return true
	}, "the server leaked search handles across repeated enumerations")
}

// TestQueryPathInformation asserts the path-based query levels, which are how a
// client inspects a file it does not open.
func TestQueryPathInformation(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	contents := []byte("twenty-four characters..")
	if err := fs.AddFile("described.txt", contents); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("adir"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	basic, err := client.GetFileBasicInfo("described.txt")
	if err != nil {
		t.Fatalf("GetFileBasicInfo() error = %v", err)
	}
	if basic.Extfileattributes == 0 {
		t.Error("the basic information reports no attributes at all")
	}

	standard, err := client.GetFileStandardInfo("described.txt")
	if err != nil {
		t.Fatalf("GetFileStandardInfo() error = %v", err)
	}
	if standard.Endoffile.QuadPart != uint64(len(contents)) {
		t.Errorf("the standard information reports %d bytes, want %d", standard.Endoffile.QuadPart, len(contents))
	}
	if standard.Directory != 0 {
		t.Error("a file is reported as a directory")
	}

	// A directory is reported as one.
	directoryInfo, err := client.GetFileStandardInfo("adir")
	if err != nil {
		t.Fatalf("GetFileStandardInfo() on a directory error = %v", err)
	}
	if directoryInfo.Directory == 0 {
		t.Error("a directory is not reported as one")
	}

	// A path that does not exist is refused.
	if _, err := client.GetFileBasicInfo("missing.txt"); err == nil {
		t.Error("GetFileBasicInfo() succeeded for a path that does not exist")
	}

	// A level that is not served is refused rather than answered with the nearest
	// one, since a client parses what it asked for.
	if _, err := client.QueryPathInformation("described.txt", 0x0999); err == nil {
		t.Error("QueryPathInformation() succeeded for an unserved information level")
	}
}

// TestQueryFileInformation asserts the handle-based query levels.
func TestQueryFileInformation(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("open.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("open.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer client.CloseFile(fid)

	data, err := client.QueryFileInformation(fid, smb1client.InfoLevelQueryFileStandard)
	if err != nil {
		t.Fatalf("QueryFileInformation() error = %v", err)
	}
	if len(data) < 24 {
		t.Fatalf("the standard information is %d bytes, want at least 24", len(data))
	}

	// A handle that is not open is refused.
	if _, err := client.QueryFileInformation(0x7FFF, smb1client.InfoLevelQueryFileStandard); err == nil {
		t.Error("QueryFileInformation() succeeded on a handle that was never issued")
	}
}

// TestQueryFsInformation asserts the volume query levels, which a client uses to
// report free space and the file-system name.
func TestQueryFsInformation(t *testing.T) {
	fs := NewMemoryFileSystem("VOLUME")
	_, client := fileServer(t, fs, false)

	size, err := client.GetFsSizeInfo()
	if err != nil {
		t.Fatalf("GetFsSizeInfo() error = %v", err)
	}
	if size.Bytespersector == 0 || size.Sectorsperallocationunit == 0 {
		t.Error("the size information reports a zero allocation geometry")
	}
	if size.Totalallocationunits.QuadPart == 0 {
		t.Error("the size information reports no capacity at all")
	}

	attribute, err := client.GetFsAttributeInfo()
	if err != nil {
		t.Fatalf("GetFsAttributeInfo() error = %v", err)
	}
	if attribute.Maxfilenamelengthinbytes == 0 {
		t.Error("the attribute information reports a zero maximum name length")
	}

	// An unserved level is refused.
	if _, err := client.QueryFsInformation(0x0999); err == nil {
		t.Error("QueryFsInformation() succeeded for an unserved information level")
	}
}

// TestSetFileInformation asserts the set levels: resizing a file, and marking one
// for deletion on close.
func TestSetFileInformation(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("resized.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddFile("doomed.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	// Resize through a handle.
	fid, err := client.OpenFile("resized.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if err := client.SetFileEndOfFile(fid, 4); err != nil {
		t.Fatalf("SetFileEndOfFile() error = %v", err)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
	attr, err := fs.Stat("resized.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if attr.Size != 4 {
		t.Fatalf("the file is %d bytes after being resized to 4", attr.Size)
	}

	// Delete-on-close is a property of the handle, and takes effect when it
	// closes rather than when it is set.
	doomed, err := client.OpenFile("doomed.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if err := client.SetFileDeleteOnClose(doomed, true); err != nil {
		t.Fatalf("SetFileDeleteOnClose() error = %v", err)
	}
	if _, err := fs.Stat("doomed.txt"); err != nil {
		t.Fatal("the file was deleted when delete-on-close was set rather than when the handle closed")
	}
	if err := client.CloseFile(doomed); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
	if _, err := fs.Stat("doomed.txt"); err == nil {
		t.Fatal("the file survived a close with delete-on-close set")
	}
}

// TestTransaction2RefusesTraversal asserts the containment boundary holds through
// the transaction subcommands too, which take a path like any other command.
func TestTransaction2RefusesTraversal(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("inside.txt", []byte("in")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	escapes := []string{
		"..\\*",
		"..\\..\\etc\\*",
		"C:\\windows\\*",
		"\\dir\\..\\..\\*",
	}
	for _, pattern := range escapes {
		pattern := pattern
		t.Run("list "+pattern, func(t *testing.T) {
			if entries, err := client.ListEntries(pattern); err == nil {
				t.Fatalf("ListEntries(%q) succeeded and returned %d entries", pattern, len(entries))
			}
		})
	}

	for _, path := range []string{"..\\outside.txt", "C:\\windows\\win.ini", "inside.txt:stream"} {
		path := path
		t.Run("query "+path, func(t *testing.T) {
			if _, err := client.GetFileBasicInfo(path); err == nil {
				t.Fatalf("GetFileBasicInfo(%q) succeeded", path)
			}
		})
	}
}

// TestSetPathInformationOnReadOnlyShare asserts a read-only share refuses the set
// levels, since they modify.
func TestSetPathInformationOnReadOnlyShare(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("guarded.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, true)

	// Reading about the file is fine.
	if _, err := client.GetFileStandardInfo("guarded.txt"); err != nil {
		t.Fatalf("GetFileStandardInfo() on a read-only share error = %v", err)
	}

	// Changing it is not.
	if err := client.SetPathInformation("guarded.txt", smb1client.InfoLevelSetFileEndOfFile,
		make([]byte, 8)); err == nil {
		t.Fatal("SetPathInformation() succeeded on a read-only share")
	}

	attr, err := fs.Stat("guarded.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if attr.Size != 10 {
		t.Fatalf("the file is %d bytes, so the refused change still landed", attr.Size)
	}
}

// TestTransaction2FragmentPlacement covers the reassembly arithmetic directly,
// including the inconsistencies it has to refuse. A fragment landing in the wrong
// place would corrupt a request in a way that is very hard to see from the outside.
func TestTransaction2FragmentPlacement(t *testing.T) {
	newReassembly := func(parameters, data int) *transactionReassembly {
		return &transactionReassembly{
			parameters: make([]byte, parameters),
			data:       make([]byte, data),
		}
	}

	t.Run("fragments assemble in order", func(t *testing.T) {
		r := newReassembly(6, 0)
		if err := r.place([]byte{1, 2, 3}, 0, nil, 0); err != nil {
			t.Fatalf("place() error = %v", err)
		}
		if r.complete() {
			t.Fatal("the transaction reports complete with half its parameters")
		}
		if err := r.place([]byte{4, 5, 6}, 3, nil, 0); err != nil {
			t.Fatalf("place() error = %v", err)
		}
		if !r.complete() {
			t.Fatal("the transaction does not report complete")
		}
		for i, want := range []byte{1, 2, 3, 4, 5, 6} {
			if r.parameters[i] != want {
				t.Fatalf("parameters = %v, want 1..6", r.parameters)
			}
		}
	})

	t.Run("fragments assemble out of order", func(t *testing.T) {
		r := newReassembly(4, 0)
		if err := r.place([]byte{3, 4}, 2, nil, 0); err != nil {
			t.Fatalf("place() error = %v", err)
		}
		if err := r.place([]byte{1, 2}, 0, nil, 0); err != nil {
			t.Fatalf("place() error = %v", err)
		}
		if !r.complete() {
			t.Fatal("the transaction does not report complete")
		}
		for i, want := range []byte{1, 2, 3, 4} {
			if r.parameters[i] != want {
				t.Fatalf("parameters = %v, want 1..4", r.parameters)
			}
		}
	})

	t.Run("inconsistent fragments are refused", func(t *testing.T) {
		cases := []struct {
			name         string
			blockSize    int
			run          []byte
			displacement int
		}{
			{"displacement past the block", 4, []byte{1}, 5},
			{"negative displacement", 4, []byte{1}, -1},
			{"run overruns the block", 4, []byte{1, 2, 3}, 2},
			{"run larger than the whole block", 2, []byte{1, 2, 3, 4}, 0},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				r := newReassembly(tc.blockSize, 0)
				if err := r.place(tc.run, tc.displacement, nil, 0); err == nil {
					t.Fatal("place() accepted an inconsistent fragment")
				}
			})
		}
	})
}

// containsName reports whether a listing holds a name.
func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

// pad4 renders an index as four digits, so generated names sort predictably.
func pad4(value int) string {
	digits := []byte{'0', '0', '0', '0'}
	for i := 3; i >= 0 && value > 0; i-- {
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits)
}

// TestFindNextRefusesAForeignTree asserts a search can only be continued through
// the tree it was opened on.
//
// FIND_NEXT2 compared the request's TID against the value the search had recorded,
// which skipped anyTreeFor and with it the check that the session owning the TID
// is the one making the request. It also compared identifiers rather than trees,
// and a TID is released for reuse when its tree disconnects while the search is
// not, so a stale search's TID could come to name a live tree on another share.
func TestFindNextRefusesAForeignTree(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("entry.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	srv, _ := fileServer(t, fs, false)

	share := &Share{Name: "other", Type: ShareTypeDisk, FS: fs}
	if err := srv.AddShare(share); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	newConnection := func() *Connection {
		return &Connection{
			Server:   srv,
			sessions: map[uint16]*Session{},
			trees:    map[uint16]*Tree{},
			tids:     newIdentifierAllocator(16),
			opens:    map[uint16]*Open{},
			fids:     newIdentifierAllocator(16),
			searches: map[uint16]*Search{},
			sids:     newIdentifierAllocator(16),
		}
	}

	// A search opened on one tree, by the session that owns it.
	conn := newConnection()
	conn.addSession(&Session{UID: 1})
	conn.addSession(&Session{UID: 2})

	tid, err := conn.tids.Allocate()
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}
	tree := &Tree{TID: tid, Share: share, SessionUID: 1}
	conn.addTree(tree)

	sid, err := conn.sids.Allocate()
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}
	conn.searches[sid] = &Search{
		SID:     sid,
		Tree:    tree,
		Entries: []DirEntry{{Attr: FileAttr{Name: "entry.txt"}}},
	}

	continueSearch := func(uid, tid uint16) nt_status.NT_STATUS {
		t.Helper()
		parameters := make([]byte, 12)
		binary.LittleEndian.PutUint16(parameters[0:2], sid)
		binary.LittleEndian.PutUint16(parameters[2:4], 10)
		binary.LittleEndian.PutUint16(parameters[4:6], 0) // InformationLevel
		request := newRequest(codes.SMB_COM_TRANSACTION2)
		request.Header.UID = types.USHORT(uid)
		request.Header.TID = types.USHORT(tid)
		_, _, status := handleFindNext2(conn, request, &transactionReassembly{parameters: parameters})
		return status
	}

	// The session that does not own the tree cannot continue the search, even
	// though it names the right TID.
	if status := continueSearch(2, tid); status == nt_status.NT_STATUS_SUCCESS {
		t.Error("a session that does not own the tree continued its search")
	}

	// Disconnecting the tree frees the TID, and the next tree takes it. The stale
	// search must not follow the identifier onto the new tree.
	conn.removeTree(tid)
	reused, err := conn.tids.Allocate()
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}
	if reused != tid {
		t.Skipf("the allocator did not reuse TID 0x%04X, so the reuse case is not exercised", tid)
	}
	conn.addTree(&Tree{TID: reused, Share: share, SessionUID: 2})

	if _, stillOpen := conn.searches[sid]; stillOpen {
		if status := continueSearch(2, reused); status == nt_status.NT_STATUS_SUCCESS {
			t.Error("a search continued through a TID that had been reallocated to another tree")
		}
	}
}
