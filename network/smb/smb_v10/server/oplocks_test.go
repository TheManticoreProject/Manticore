package server

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// oplockServer stands up a server with one memory share holding a file, and
// returns the server and a client connected to it.
func oplockServer(t *testing.T) (*Server, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("cached.bin", []byte("the bytes a client would cache")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.Mkdir("folder"); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	return fileServer(t, fs, false)
}

// attachClient connects a second client to a server already running, so a break
// crossing from one connection to another can be observed.
func attachClient(t *testing.T, srv *Server) *smb1client.Client {
	t.Helper()

	serverSide, clientSide := net.Pipe()
	remote := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2}

	served := make(chan struct{})
	go func() {
		defer close(served)
		_ = srv.ServeConn(tcp.NewTCPTransportFromConn(serverSide), remote)
	}()

	transportEnd := tcp.NewTCPTransportFromConn(clientSide)
	transportEnd.SetTimeout(5 * time.Second)
	t.Cleanup(func() {
		clientSide.Close()
		serverSide.Close()
		<-served
	})

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect(fileShareName); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}
	return client
}

// openAsking opens a file with a chosen NT_CREATE_ANDX Flags field and returns
// the handle and the oplock level granted.
//
// The client API does not carry the Flags field, and the oplock request lives
// there, so the request is built here.
func openAsking(
	t *testing.T,
	client *smb1client.Client,
	name string,
	createFlags uint32,
	access uint32,
	options uint32,
) (smb1client.FID, uint8) {
	t.Helper()

	request := newRequest(codes.SMB_COM_NT_CREATE_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= 0x8000

	cmd := commands.NewNtCreateAndxRequest()
	cmd.FileName = *types.NewSMB_STRING([]byte("\\" + name))
	cmd.Flags = types.ULONG(createFlags)
	cmd.DesiredAccess = types.ULONG(access)
	cmd.ShareAccess = types.ULONG(fileflags.FILE_SHARE_READ | fileflags.FILE_SHARE_WRITE | fileflags.FILE_SHARE_DELETE)
	cmd.CreateDisposition = types.ULONG(fileflags.FILE_OPEN)
	cmd.CreateOptions = types.ULONG(options)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the open: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Fatalf("opening %q reported 0x%08X, want success", name, status)
	}

	response := commands.NewNtCreateAndxResponse()
	if _, err := response.Unmarshal(raw[header.SMB_HEADER_SIZE:]); err != nil {
		t.Fatalf("the open reply did not decode: %v", err)
	}
	return smb1client.FID(response.FID), uint8(response.OpLockLevel)
}

// awaitBreak reads the next frame on a client's transport and decodes it as an
// oplock break notification.
//
// A break is a request rather than a response, so a client that is not waiting
// for anything still has one arrive on its transport. That is what this reads.
func awaitBreak(t *testing.T, client *smb1client.Client) *commands.LockingAndxRequest {
	t.Helper()

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("no oplock break arrived: %v", err)
	}

	decoded := message.NewMessage()
	if err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("the break did not decode: %v", err)
	}
	if decoded.Header.Command != codes.SMB_COM_LOCKING_ANDX {
		t.Fatalf("the server sent command 0x%02X, want SMB_COM_LOCKING_ANDX",
			uint8(decoded.Header.Command))
	}
	// A break is a request: [MS-CIFS] section 3.3.4.1 has SMB_FLAGS_REPLY set on
	// every message the server sends "unless the message is an OpLock Break
	// Notification request initiated by the server".
	if decoded.Header.Flags.IsReply() {
		t.Error("the break has SMB_FLAGS_REPLY set, so a client would decode it as a response")
	}

	notification, ok := decoded.Command.(*commands.LockingAndxRequest)
	if !ok {
		t.Fatalf("the break decoded as %T, want a LockingAndxRequest", decoded.Command)
	}
	if uint8(notification.TypeOfLock)&commands.LockingAndxOplockRelease == 0 {
		t.Errorf("the break's TypeOfLock is 0x%02X, want the OPLOCK_RELEASE bit (0x%02X)",
			notification.TypeOfLock, commands.LockingAndxOplockRelease)
	}
	if notification.NewOpLockLevel != breakToNone {
		t.Errorf("the break reports NewOpLockLevel %d, want 0 — level II breaks to nothing",
			notification.NewOpLockLevel)
	}
	return notification
}

// expectNoBreak asserts nothing arrives on a client's transport.
func expectNoBreak(t *testing.T, client *smb1client.Client) {
	t.Helper()

	client.Transport.SetTimeout(250 * time.Millisecond)
	defer client.Transport.SetTimeout(5 * time.Second)

	if raw, err := client.Transport.Receive(); err == nil {
		t.Errorf("an unexpected %d-byte message arrived, first bytes % x", len(raw), raw[:min(16, len(raw))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestOplockGrantedOnAnOpenThatAsksForOne(t *testing.T) {
	_, client := oplockServer(t)

	_, level := openAsking(t, client, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)

	if level != OplockLevelLevelII {
		t.Errorf("the open was granted oplock level %d, want level II (%d)", level, OplockLevelLevelII)
	}
}

func TestBatchOplockRequestIsDowngradedToLevelII(t *testing.T) {
	// [MS-SMB] section 3.3.5.1.2 sanctions answering a request for an exclusive or
	// batch oplock with level II. Refusing outright would leave a client that asks
	// for a batch oplock with no caching at all, which the capability exists to
	// avoid.
	_, client := oplockServer(t)

	for _, asked := range []uint32{ntCreateRequestOplock, ntCreateRequestOpBatch, ntCreateRequestOplock | ntCreateRequestOpBatch} {
		_, level := openAsking(t, client, "cached.bin", asked,
			fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)
		if level != OplockLevelLevelII {
			t.Errorf("an open asking with flags 0x%08X was granted level %d, want level II (%d)",
				asked, level, OplockLevelLevelII)
		}
	}
}

func TestOplockNotGrantedWhenNoneIsAskedFor(t *testing.T) {
	// A client that did not ask has not agreed to answer a break, so granting one
	// would have the server expect an acknowledgement it will never get and the
	// client cache on a promise it does not know it holds.
	_, client := oplockServer(t)

	_, level := openAsking(t, client, "cached.bin", 0,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)

	if level != OplockLevelNone {
		t.Errorf("an open that asked for nothing was granted level %d, want none", level)
	}
}

func TestOplockNotGrantedOnADirectory(t *testing.T) {
	// "If the open or create is on a directory file, then an Oplock MUST NOT be
	// granted" ([MS-CIFS] section 3.3.5.2.7).
	_, client := oplockServer(t)

	_, level := openAsking(t, client, "folder", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_DIRECTORY_FILE)

	if level != OplockLevelNone {
		t.Errorf("an open of a directory was granted level %d, want none", level)
	}
}

func TestOplockNotGrantedWithASingleCommandSlot(t *testing.T) {
	// [MS-CIFS] section 3.3.5.53: a MaxMpxCount below two turns oplock support off
	// for the connection, "because there would not be enough outstanding command
	// slots to properly revoke the OpLock". Asserted against the decision rather
	// than over the wire, because the client fixes its own MaxMpxCount from what
	// the server advertised.
	conn := &Connection{
		Server:            &Server{config: Config{}},
		Negotiated:        true,
		ClientMaxMpxCount: 1,
	}
	open := &Open{Path: "cached.bin"}

	if got := conn.grantOplock(open, 1, ntCreateRequestOplock); got != OplockLevelNone {
		t.Errorf("a client with one command slot was granted level %d, want none", got)
	}

	// And two slots is enough, given a share to record it on.
	conn.ClientMaxMpxCount = 2
	share := &Share{Name: "files", oplocks: newOplockTable()}
	open.Tree = &Tree{TID: 1, Share: share}
	if got := conn.grantOplock(open, 1, ntCreateRequestOplock); got != OplockLevelLevelII {
		t.Errorf("a client with two command slots was granted level %d, want level II", got)
	}
}

func TestOplockBrokenByAWriteFromAnotherConnection(t *testing.T) {
	// The exchange the capability exists for: one client caches on a level II
	// oplock, another writes, and the first is told before the bytes land.
	srv, holder := oplockServer(t)
	writer := attachClient(t, srv)

	_, level := openAsking(t, holder, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)
	if level != OplockLevelLevelII {
		t.Fatalf("the holder was granted level %d, want level II", level)
	}

	// The writer's own open breaks nothing yet — it has to be able to write for
	// that — so open it read-write and then write through it.
	fid, err := writer.OpenFile("cached.bin",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("the writer could not open the file: %v", err)
	}

	// The writable open is itself a break, so the notification is already on its
	// way before the write.
	awaitBreak(t, holder)

	if _, err := writer.WriteFile(fid, 0, []byte("replaced")); err != nil {
		t.Fatalf("the write failed: %v", err)
	}

	// And the oplock is gone rather than broken twice: a second break would have
	// the client acknowledge something it no longer holds.
	expectNoBreak(t, holder)

	share := srv.Share(fileShareName)
	if held := share.oplocks.HeldOn("cached.bin"); held != 0 {
		t.Errorf("%d oplocks are still recorded on the file, want none", held)
	}
}

func TestOplockBrokenByAWriteThroughAHandleOpenedBeforeIt(t *testing.T) {
	// A writer that was already there when the oplock was granted. [MS-CIFS]
	// section 3.3.5.2.7 would have refused the oplock; this grants it and
	// withdraws it when the writer actually writes, which keeps the same promise.
	srv, writer := oplockServer(t)
	holder := attachClient(t, srv)

	fid, err := writer.OpenFile("cached.bin",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("the writer could not open the file: %v", err)
	}

	_, level := openAsking(t, holder, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)
	if level != OplockLevelLevelII {
		t.Fatalf("the holder was granted level %d, want level II", level)
	}

	if _, err := writer.WriteFile(fid, 0, []byte("replaced")); err != nil {
		t.Fatalf("the write failed: %v", err)
	}

	awaitBreak(t, holder)
}

func TestOplockBrokenByADelete(t *testing.T) {
	// A delete is reached by path, so no open of the deleting client has broken
	// anything. Without an explicit break the holder would go on caching a file
	// that no longer exists.
	srv, holder := oplockServer(t)
	deleter := attachClient(t, srv)

	if _, level := openAsking(t, holder, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE); level != OplockLevelLevelII {
		t.Fatalf("the holder was granted level %d, want level II", level)
	}

	if err := deleter.DeleteFile("cached.bin"); err != nil {
		t.Fatalf("the delete failed: %v", err)
	}

	awaitBreak(t, holder)
}

func TestOplockBrokenByARename(t *testing.T) {
	srv, holder := oplockServer(t)
	renamer := attachClient(t, srv)

	if _, level := openAsking(t, holder, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE); level != OplockLevelLevelII {
		t.Fatalf("the holder was granted level %d, want level II", level)
	}

	if err := renamer.RenameFile("cached.bin", "moved.bin"); err != nil {
		t.Fatalf("the rename failed: %v", err)
	}

	awaitBreak(t, holder)
}

func TestOplockNotBrokenByAnotherReader(t *testing.T) {
	// Level II is "multiple readers of a file and no writers", so a second reader
	// is exactly what it is for. Breaking on one would make the oplock useless.
	srv, holder := oplockServer(t)
	reader := attachClient(t, srv)

	if _, level := openAsking(t, holder, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE); level != OplockLevelLevelII {
		t.Fatalf("the holder was granted level %d, want level II", level)
	}

	fid, err := reader.OpenFile("cached.bin",
		fileflags.GENERIC_READ,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("the second reader could not open the file: %v", err)
	}
	if _, err := reader.ReadFile(fid, 0, 8); err != nil {
		t.Fatalf("the second reader could not read: %v", err)
	}

	expectNoBreak(t, holder)

	share := srv.Share(fileShareName)
	if held := share.oplocks.HeldOn("cached.bin"); held != 1 {
		t.Errorf("%d oplocks are recorded on the file, want the holder's one", held)
	}
}

func TestOplockReleasedByTheClientsAcknowledgement(t *testing.T) {
	// [MS-CIFS] section 3.3.5.30: the release sets the handle's oplock to NONE and
	// is answered with silence, since both range counts are zero.
	srv, client := oplockServer(t)

	fid, level := openAsking(t, client, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)
	if level != OplockLevelLevelII {
		t.Fatalf("the open was granted level %d, want level II", level)
	}

	share := srv.Share(fileShareName)
	if held := share.oplocks.HeldOn("cached.bin"); held != 1 {
		t.Fatalf("%d oplocks are recorded, want one", held)
	}

	request := newRequest(codes.SMB_COM_LOCKING_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	release := commands.NewLockingAndxRequest()
	release.FID = types.USHORT(fid)
	release.TypeOfLock = types.UCHAR(commands.LockingAndxOplockRelease)
	request.AddCommand(release)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the release: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Answered with silence, so the only way to see it landed is the table. An
	// echo afterwards proves the connection is still in step rather than holding
	// an unsent response.
	waitFor(t, func() bool { return share.oplocks.HeldOn("cached.bin") == 0 },
		"the release did not drop the oplock")
	if _, err := client.Echo([]byte("still here")); err != nil {
		t.Errorf("the connection did not survive the release: %v", err)
	}
}

func TestOplockReleasedWhenTheHandleCloses(t *testing.T) {
	srv, client := oplockServer(t)

	fid, level := openAsking(t, client, "cached.bin", ntCreateRequestOplock,
		fileflags.GENERIC_READ, fileflags.FILE_NON_DIRECTORY_FILE)
	if level != OplockLevelLevelII {
		t.Fatalf("the open was granted level %d, want level II", level)
	}

	share := srv.Share(fileShareName)
	if held := share.oplocks.HeldOn("cached.bin"); held != 1 {
		t.Fatalf("%d oplocks are recorded, want one", held)
	}

	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("closing the handle failed: %v", err)
	}
	if held := share.oplocks.HeldOn("cached.bin"); held != 0 {
		t.Errorf("%d oplocks are recorded after the close, want none", held)
	}
}

func TestOpenReportsTheOplockItHolds(t *testing.T) {
	// Open.OplockLevel asks the table rather than keeping a copy, so it cannot
	// disagree with what a break has already withdrawn — and cannot be written
	// from two goroutines at once, which a field on the handle would be.
	table := newOplockTable()
	share := &Share{Name: "files", oplocks: table}
	tree := &Tree{TID: 1, Share: share}
	open := &Open{FID: 1, Path: "cached.bin", Tree: tree}

	if got := open.OplockLevel(); got != OplockLevelNone {
		t.Errorf("a handle with no oplock reports level %d, want none", got)
	}

	if !table.Grant(open, &Connection{}, 1) {
		t.Fatal("granting failed")
	}
	if got := open.OplockLevel(); got != OplockLevelLevelII {
		t.Errorf("a handle holding an oplock reports level %d, want level II", got)
	}

	table.Break("cached.bin", nil)
	if got := open.OplockLevel(); got != OplockLevelNone {
		t.Errorf("a handle whose oplock was broken reports level %d, want none", got)
	}

	// A handle on a share with no table, and a nil handle, report none rather
	// than panicking: every mutation path asks.
	if got := (&Open{Path: "x", Tree: &Tree{Share: &Share{}}}).OplockLevel(); got != OplockLevelNone {
		t.Errorf("a handle on a share with no table reports level %d, want none", got)
	}
	var missing *Open
	if got := missing.OplockLevel(); got != OplockLevelNone {
		t.Errorf("a nil handle reports level %d, want none", got)
	}
}

func TestCapabilityIsAdvertised(t *testing.T) {
	// The capability is a promise that a break can be sent, so it belongs in the
	// negotiate response only now that one can.
	if serverCapabilities&0x00000080 == 0 {
		t.Error("CAP_LEVEL_II_OPLOCKS is not advertised, so no client will ask for an oplock")
	}
}

func TestOplockTableBreaksEveryHolderButOne(t *testing.T) {
	// The table itself, since the wire tests can only reach it two handles at a
	// time.
	table := newOplockTable()
	share := &Share{Name: "files", oplocks: table}
	tree := &Tree{TID: 1, Share: share}

	first := &Open{FID: 1, Path: "a/b.txt", Tree: tree}
	second := &Open{FID: 2, Path: "A/B.TXT", Tree: tree}
	third := &Open{FID: 3, Path: "a/b.txt", Tree: tree}
	other := &Open{FID: 4, Path: "a/c.txt", Tree: tree}

	conn := &Connection{}
	for _, open := range []*Open{first, second, third, other} {
		if !table.Grant(open, conn, 1) {
			t.Fatalf("granting on FID %d failed", open.FID)
		}
	}

	// The three spellings of one path share an entry, because the file commands
	// match names case-insensitively and a break must not miss a holder over
	// capitalisation.
	if held := table.HeldOn("a/b.txt"); held != 3 {
		t.Errorf("%d oplocks are recorded on the file, want 3 across its spellings", held)
	}

	broken := table.Break("A/b.TXT", second)
	if len(broken) != 2 {
		t.Fatalf("breaking left %d holders broken, want 2", len(broken))
	}
	for _, holder := range broken {
		if holder.owner == second {
			t.Error("the excepted handle was broken")
		}
	}
	if held := table.HeldOn("a/b.txt"); held != 1 {
		t.Errorf("%d oplocks remain, want the excepted one", held)
	}
	if held := table.HeldOn("a/c.txt"); held != 1 {
		t.Errorf("breaking one file dropped %d oplocks on another", 1-held)
	}

	// Releasing the last one empties the entry, and releasing again reports that
	// nothing was held rather than failing.
	if !table.Release(second) {
		t.Error("releasing the remaining holder reported it held nothing")
	}
	if table.Release(second) {
		t.Error("releasing twice reported an oplock the second time")
	}
	if held := table.HeldOn("a/b.txt"); held != 0 {
		t.Errorf("%d oplocks remain after the release, want none", held)
	}
}

func TestBreakingWithNoHoldersIsHarmless(t *testing.T) {
	// Every mutation path calls the break, and almost none of them will find a
	// holder, so the empty case has to cost nothing and touch nothing.
	table := newOplockTable()
	if broken := table.Break("nothing/here.txt", nil); len(broken) != 0 {
		t.Errorf("breaking an unheld file reported %d holders", len(broken))
	}

	conn := &Connection{Server: &Server{config: Config{}}}
	conn.breakOplocksOn(nil, "nothing/here.txt", nil)
	conn.breakOplocksOn(&Share{Name: "files"}, "nothing/here.txt", nil)
}

func TestUnsolicitedMIDsAreDistinctAndNonZero(t *testing.T) {
	// A client matches a response to a request by MID, so two breaks in flight
	// must not share one. Zero is skipped because a client may read it as "no
	// MID".
	conn := &Connection{}

	seen := map[uint16]bool{}
	for round := 0; round < 4; round++ {
		mid := conn.nextUnsolicitedMID()
		if mid == 0 {
			t.Fatal("an unsolicited MID of zero was handed out")
		}
		if seen[mid] {
			t.Fatalf("MID 0x%04X was handed out twice", mid)
		}
		seen[mid] = true
	}
}
