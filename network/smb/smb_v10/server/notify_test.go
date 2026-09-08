package server

import (
	"encoding/binary"
	"testing"
	"time"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// notifyReply is what a watch's answer carried.
type notifyReply struct {
	status  uint32
	records []filesystem.FileNotifyInformation
}

// watchDirectory opens a directory handle and sends NT_TRANSACT_NOTIFY_CHANGE on
// it, returning a channel that receives the answer whenever it arrives.
//
// The reply is read on a goroutine because the whole point of the command is that
// it is answered later: reading inline would block the test before it could make
// the change it wants to be told about.
func watchDirectory(
	t *testing.T,
	client *smb1client.Client,
	path string,
	completionFilter uint32,
	watchTree bool,
	maxParameterCount uint32,
) <-chan notifyReply {
	t.Helper()

	fid, err := client.OpenFile(path,
		fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening %q as a directory failed: %v", path, err)
	}

	tree := uint16(0)
	if watchTree {
		tree = 1
	}
	setup := []types.USHORT{
		types.USHORT(completionFilter & 0xFFFF),
		types.USHORT(completionFilter >> 16),
		types.USHORT(fid),
		types.USHORT(tree),
	}

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(4) // NT_TRANSACT_NOTIFY_CHANGE
	transaction.Setup = setup
	transaction.SetupCount = types.UCHAR(len(setup))
	transaction.MaxParameterCount = types.ULONG(maxParameterCount)
	transaction.MaxDataCount = types.ULONG(0)
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the watch: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	replies := make(chan notifyReply, 1)
	go func() {
		raw, err := client.Transport.Receive()
		if err != nil {
			replies <- notifyReply{status: 0xFFFFFFFF}
			return
		}
		if len(raw) < 9 {
			replies <- notifyReply{status: 0xFFFFFFFE}
			return
		}

		reply := notifyReply{status: binary.LittleEndian.Uint32(raw[5:9])}
		if reply.status == 0 {
			reply.records = notifyRecordsOf(raw)
		}
		replies <- reply
	}()
	return replies
}

// notifyRecordsOf pulls the FILE_NOTIFY_INFORMATION records out of a reply, using
// the offsets the reply itself declares.
//
// They are parsed with windows/filesystem's own reader rather than by hand, so a
// layout this server got wrong fails here instead of agreeing with itself.
func notifyRecordsOf(raw []byte) []filesystem.FileNotifyInformation {
	// WordCount(1) then the response words; ParameterCount is the fourth 4-byte
	// word after the 3 reserved bytes, and ParameterOffset the fifth.
	words := raw[header.SMB_HEADER_SIZE+1:]
	if len(words) < 19 {
		return nil
	}
	count := int(binary.LittleEndian.Uint32(words[11:15]))
	offset := int(binary.LittleEndian.Uint32(words[15:19]))

	if count == 0 || offset+count > len(raw) {
		return nil
	}
	return filesystem.ParseFileNotifyInformation(raw[offset : offset+count])
}

// notifyServer serves a share with a directory to watch.
func notifyServer(t *testing.T) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddDirectory("watched"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)
	return fs, client
}

// TestNotifyChangeAnswersWhenSomethingIsCreated asserts a watch is answered when
// an entry appears, with the name and action of the change.
//
// This is the command the deferred response path was built for: it is asked before
// there is anything to say and answered when there is.
func TestNotifyChangeAnswersWhenSomethingIsCreated(t *testing.T) {
	fs, client := notifyServer(t)

	replies := watchDirectory(t, client, "watched", 0, false, 4096)

	// The change is made through the backend rather than over the wire, because
	// the connection is occupied waiting for the answer.
	waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")
	if err := fs.AddFile(`watched/appeared.txt`, []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	select {
	case reply := <-replies:
		if reply.status != 0 {
			t.Fatalf("the watch reported 0x%08X, want success", reply.status)
		}
		if len(reply.records) != 1 {
			t.Fatalf("the answer carries %d records, want 1", len(reply.records))
		}
		if got := reply.records[0].FileName; got != "appeared.txt" {
			t.Errorf("the change names %q, want %q", got, "appeared.txt")
		}
		if reply.records[0].Action != filesystem.FileActionAdded {
			t.Errorf("the action is %v, want added", reply.records[0].Action)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the watch was never answered")
	}
}

// TestNotifyChangeTellsTheClientToReEnumerateWhenItWillNotFit asserts a client
// whose buffer is too small is told to list the directory again.
//
// [MS-CIFS] section 2.2.7.4: too much to return means "zero bytes are returned and
// STATUS_NOTIFY_ENUM_DIR [...] is returned in the Status field". A client handles
// that by re-enumerating, so honouring a small buffer is a complete answer rather
// than a failure — and truncating a record instead would give the client a name
// that is not a name.
func TestNotifyChangeTellsTheClientToReEnumerateWhenItWillNotFit(t *testing.T) {
	fs, client := notifyServer(t)

	// Four bytes cannot hold even the fixed part of one record.
	replies := watchDirectory(t, client, "watched", 0, false, 4)

	waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")
	if err := fs.AddFile(`watched/appeared.txt`, []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	select {
	case reply := <-replies:
		if reply.status != uint32(nt_status.NT_STATUS_NOTIFY_ENUM_DIR) {
			t.Errorf("a watch with a 4-byte buffer reported 0x%08X, want STATUS_NOTIFY_ENUM_DIR (0x%08X)",
				reply.status, uint32(nt_status.NT_STATUS_NOTIFY_ENUM_DIR))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the watch was never answered")
	}
}

// TestNotifyChangeHonoursItsCompletionFilter asserts a change the client did not
// ask about does not answer the watch, and one it did asks does.
func TestNotifyChangeHonoursItsCompletionFilter(t *testing.T) {
	fs, client := notifyServer(t)

	// Asking only about size changes: a new name must not answer this.
	replies := watchDirectory(t, client, "watched", filesystem.FILE_NOTIFY_CHANGE_SIZE, false, 4096)

	waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")
	if err := fs.AddFile(`watched/appeared.txt`, []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	select {
	case reply := <-replies:
		t.Fatalf("a name change answered a size-only watch, with status 0x%08X", reply.status)
	case <-time.After(300 * time.Millisecond):
		// Correct: the filter excluded it.
	}

	// A modification is what the filter admits.
	if err := fs.SetAttr(`watched/appeared.txt`, FileAttr{ReadOnly: true}, AttrMask{ReadOnly: true}); err != nil {
		t.Fatalf("SetAttr() error = %v", err)
	}

	select {
	case reply := <-replies:
		if reply.status != 0 {
			t.Fatalf("the watch reported 0x%08X, want success", reply.status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a change the filter admits did not answer the watch")
	}
}

// TestNotifyChangeOnANonDirectoryIsRefused asserts the handle has to name a
// directory.
//
// [MS-CIFS] section 2.2.7.4: "A directory file MUST be opened before this command
// can be used." A handle on a file has no entries whose changes could be reported.
func TestNotifyChangeOnANonDirectoryIsRefused(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("afile.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("afile.txt",
		fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}

	setup := []types.USHORT{0, 0, types.USHORT(fid), 0}
	status := sendNtTransact(t, client, 4, setup, nil, nil)
	if status != uint32(nt_status.NT_STATUS_NOT_A_DIRECTORY) {
		t.Errorf("watching a file reported 0x%08X, want STATUS_NOT_A_DIRECTORY (0x%08X)",
			status, uint32(nt_status.NT_STATUS_NOT_A_DIRECTORY))
	}
}

// TestNotifyChangeOnAnUnwatchableBackendIsRefused asserts a share whose backend
// cannot be watched says so.
//
// Answering success and never notifying would leave the client waiting for a
// change it would never hear about — a worse outcome than a refusal it can act on.
func TestNotifyChangeOnAnUnwatchableBackendIsRefused(t *testing.T) {
	fs := &unwatchableFileSystem{MemoryFileSystem: NewMemoryFileSystem("FILES")}
	if err := fs.AddDirectory("watched"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("watched",
		fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the directory failed: %v", err)
	}

	setup := []types.USHORT{0, 0, types.USHORT(fid), 0}
	status := sendNtTransact(t, client, 4, setup, nil, nil)
	if status != uint32(nt_status.NT_STATUS_NOT_SUPPORTED) {
		t.Errorf("watching an unwatchable share reported 0x%08X, want STATUS_NOT_SUPPORTED", status)
	}
}

// unwatchableFileSystem is a MemoryFileSystem with Watch hidden, standing in for a
// caller's backend that cannot report changes.
type unwatchableFileSystem struct {
	*MemoryFileSystem

	// Watch shadows the embedded method, so the type does not satisfy Watcher.
	Watch struct{}
}

// TestNotifyChangeStopsWatchingWhenCancelled asserts a cancelled watch releases
// the backend watch rather than leaving it running for the life of the share.
func TestNotifyChangeStopsWatchingWhenCancelled(t *testing.T) {
	fs, client := notifyServer(t)

	replies := watchDirectory(t, client, "watched", 0, false, 4096)
	waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")

	// The cancel names the watch's own PID and MID, which newRequest fixes.
	request := newRequest(codes.SMB_COM_NT_CANCEL)
	request.Header.UID = client.Session.SessionUID
	request.AddCommand(commands.NewNtCancelRequest())

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the cancel: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// The backend watch goes when the request is cancelled, which is what keeps a
	// client that walks away from leaking one per abandoned watch.
	waitFor(t, func() bool { return fs.watches.count() == 0 },
		"the backend watch outlived the cancelled request")

	// And nothing was sent for it, so the reply stream stays in step: an echo
	// afterwards gets its own answer.
	select {
	case reply := <-replies:
		t.Fatalf("a cancelled watch was answered anyway, with status 0x%08X", reply.status)
	case <-time.After(300 * time.Millisecond):
	}
}
