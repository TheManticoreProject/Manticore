package server

import (
	"encoding/binary"
	"runtime"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// rawReply is the part of a response these tests look at.
type rawReply struct {
	command codes.CommandCode
	status  uint32
}

// sendRaw sends one request on the client's session and tree without reading
// anything back.
//
// The client's own request helpers read the next frame as their answer, which is
// wrong here: a closed handle's watch is answered first, ahead of the response to
// whatever closed it.
func sendRaw(t *testing.T, client *smb1client.Client, cmd command_interface.CommandInterface) {
	t.Helper()

	request := newRequest(cmd.GetCommandCode())
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal command 0x%02X: %v", uint8(cmd.GetCommandCode()), err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

// receiveRaw reads one frame and reports its command and status.
func receiveRaw(t *testing.T, client *smb1client.Client) rawReply {
	t.Helper()

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if len(raw) < 9 {
		t.Fatalf("a %d-byte reply is too short to carry a status", len(raw))
	}
	return rawReply{command: codes.CommandCode(raw[4]), status: binary.LittleEndian.Uint32(raw[5:9])}
}

// sendNotify parks an NT_TRANSACT_NOTIFY_CHANGE on a handle without waiting for
// its answer.
func sendNotify(t *testing.T, client *smb1client.Client, fid uint16) {
	t.Helper()

	setup := []types.USHORT{0, 0, types.USHORT(fid), 0}
	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(4) // NT_TRANSACT_NOTIFY_CHANGE
	transaction.Setup = setup
	transaction.SetupCount = types.UCHAR(len(setup))
	transaction.MaxParameterCount = types.ULONG(4096)
	sendRaw(t, client, transaction)
}

// openWatchedDirectory opens the test share's watched directory.
func openWatchedDirectory(t *testing.T, client *smb1client.Client) uint16 {
	t.Helper()

	fid, err := client.OpenFile("watched",
		fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the watched directory failed: %v", err)
	}
	return uint16(fid)
}

// closeRequest builds an SMB_COM_CLOSE for a handle.
func closeRequest(fid uint16) *commands.CloseRequest {
	request := commands.NewCloseRequest()
	request.FID = types.USHORT(fid)
	return request
}

// expectCleanupThen asserts the next two replies are a watch ended with
// STATUS_NOTIFY_CLEANUP, then a successful answer to the command that ended it.
func expectCleanupThen(t *testing.T, client *smb1client.Client, closing codes.CommandCode) {
	t.Helper()

	watch := receiveRaw(t, client)
	if watch.command != codes.SMB_COM_NT_TRANSACT || watch.status != uint32(nt_status.NT_STATUS_NOTIFY_CLEANUP) {
		t.Fatalf("the pending watch was answered as command 0x%02X status 0x%08X, want NT_TRANSACT with STATUS_NOTIFY_CLEANUP",
			uint8(watch.command), watch.status)
	}
	closed := receiveRaw(t, client)
	if closed.command != closing || closed.status != 0 {
		t.Fatalf("the closing command was answered as command 0x%02X status 0x%08X, want 0x%02X with success",
			uint8(closed.command), closed.status, uint8(closing))
	}
}

// onlyConnection returns the one connection a test server is serving.
func onlyConnection(t *testing.T, srv *Server) *Connection {
	t.Helper()

	srv.mutex.RLock()
	defer srv.mutex.RUnlock()
	for conn := range srv.conns {
		return conn
	}
	t.Fatal("the server is serving no connection")
	return nil
}

// TestNotifyChangeEndsWhenItsHandleIsClosed asserts that every way of closing a
// handle completes the watch pending on it with STATUS_NOTIFY_CLEANUP, ahead of
// the answer to the close, and stops the backend watch.
//
// [MS-CIFS] section 3.3.5.59.4 completes the request when "the FID is closed,
// either by an explicit Close operation or another cause"; a tree disconnect and
// a logoff close the handle as surely as a close does.
func TestNotifyChangeEndsWhenItsHandleIsClosed(t *testing.T) {
	cases := []struct {
		name    string
		command codes.CommandCode
		closing func(fid uint16) command_interface.CommandInterface
	}{
		{"close", codes.SMB_COM_CLOSE, func(fid uint16) command_interface.CommandInterface {
			return closeRequest(fid)
		}},
		{"tree disconnect", codes.SMB_COM_TREE_DISCONNECT, func(uint16) command_interface.CommandInterface {
			return commands.NewTreeDisconnectRequest()
		}},
		{"logoff", codes.SMB_COM_LOGOFF_ANDX, func(uint16) command_interface.CommandInterface {
			return commands.NewLogoffAndxRequest()
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewMemoryFileSystem("FILES")
			if err := fs.AddDirectory("watched"); err != nil {
				t.Fatalf("AddDirectory() error = %v", err)
			}
			srv, client := fileServer(t, fs, false)
			conn := onlyConnection(t, srv)

			fid := openWatchedDirectory(t, client)
			sendNotify(t, client, fid)
			waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")

			sendRaw(t, client, tc.closing(fid))
			expectCleanupThen(t, client, tc.command)

			waitFor(t, func() bool { return fs.watches.count() == 0 },
				"the backend watch outlived the handle it was watching")
			if outstanding := conn.OutstandingResponses(); outstanding != 0 {
				t.Fatalf("%d deferred answers are still outstanding after the handle closed", outstanding)
			}
		})
	}
}

// TestNotifyChangeWatchesDoNotAccumulateAcrossCloses asserts that opening a
// directory, watching it and closing it, over and over on one connection, leaves
// nothing behind.
//
// Before the watch ended with its handle, every cycle left a backend watch, a
// deferred answer and a goroutine running until the connection closed, and the
// cap on open handles did not bound them because each handle was closed.
func TestNotifyChangeWatchesDoNotAccumulateAcrossCloses(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddDirectory("watched"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	srv, client := fileServer(t, fs, false)
	conn := onlyConnection(t, srv)

	before := runtime.NumGoroutine()

	// Well past the per-connection cap, so a closed handle's watch still counted
	// against it would be refused before the loop ends.
	cycles := 3 * DefaultMaxNotifiesPerConnection
	for i := 0; i < cycles; i++ {
		fid := openWatchedDirectory(t, client)
		sendNotify(t, client, fid)
		waitFor(t, func() bool { return fs.watches.count() == 1 }, "the watch was not registered")

		sendRaw(t, client, closeRequest(fid))
		expectCleanupThen(t, client, codes.SMB_COM_CLOSE)
	}

	waitFor(t, func() bool { return fs.watches.count() == 0 },
		"backend watches outlived the handles they were watching")
	if outstanding := conn.OutstandingResponses(); outstanding != 0 {
		t.Fatalf("%d deferred answers are still outstanding after every handle closed", outstanding)
	}
	waitFor(t, func() bool { return runtime.NumGoroutine() <= before+5 },
		"watch goroutines outlived the handles they were watching")
}

// TestNotifyChangeIsCappedPerConnection asserts a connection cannot hold more
// pending watches than MaxNotifiesPerConnection, and that the slots come back
// when the watches end.
func TestNotifyChangeIsCappedPerConnection(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddDirectory("watched"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid := openWatchedDirectory(t, client)
	for i := 0; i < DefaultMaxNotifiesPerConnection; i++ {
		sendNotify(t, client, fid)
	}
	waitFor(t, func() bool { return fs.watches.count() == DefaultMaxNotifiesPerConnection },
		"the watches up to the cap were not all registered")

	// One more is refused at once, without starting a watch.
	sendNotify(t, client, fid)
	refused := receiveRaw(t, client)
	if refused.status != uint32(nt_status.NT_STATUS_INSUFFICIENT_RESOURCES) {
		t.Fatalf("a watch past the cap was answered with status 0x%08X, want STATUS_INSUFFICIENT_RESOURCES",
			refused.status)
	}
	if count := fs.watches.count(); count != DefaultMaxNotifiesPerConnection {
		t.Fatalf("a refused watch left %d backend watches, want %d", count, DefaultMaxNotifiesPerConnection)
	}

	// Closing the handle ends every watch on it, each with its own answer.
	sendRaw(t, client, closeRequest(fid))
	for i := 0; i < DefaultMaxNotifiesPerConnection; i++ {
		watch := receiveRaw(t, client)
		if watch.status != uint32(nt_status.NT_STATUS_NOTIFY_CLEANUP) {
			t.Fatalf("pending watch %d was answered with status 0x%08X, want STATUS_NOTIFY_CLEANUP", i, watch.status)
		}
	}
	if closed := receiveRaw(t, client); closed.command != codes.SMB_COM_CLOSE || closed.status != 0 {
		t.Fatalf("the close was answered as command 0x%02X status 0x%08X", uint8(closed.command), closed.status)
	}
	waitFor(t, func() bool { return fs.watches.count() == 0 }, "the closed handle's watches kept running")

	// The slots are free again.
	fid = openWatchedDirectory(t, client)
	sendNotify(t, client, fid)
	waitFor(t, func() bool { return fs.watches.count() == 1 }, "a watch was refused after the cap freed up")
}

// TestNewServerRejectsANegativeNotifyCap asserts the watch cap is validated the
// way the other per-connection limits are, and defaulted when left zero.
func TestNewServerRejectsANegativeNotifyCap(t *testing.T) {
	if _, err := NewServer(Config{MaxNotifiesPerConnection: -1}); err == nil {
		t.Fatal("NewServer() accepted a negative MaxNotifiesPerConnection")
	}

	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if srv.config.MaxNotifiesPerConnection != DefaultMaxNotifiesPerConnection {
		t.Fatalf("MaxNotifiesPerConnection defaulted to %d, want %d",
			srv.config.MaxNotifiesPerConnection, DefaultMaxNotifiesPerConnection)
	}
}
