package server

import (
	"bytes"
	"encoding/binary"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// writeAndxRequestDataOffset is where a WriteAndx request's data block begins
// when the command is first in the message, measured from the start of the SMB
// header: header + WordCount(1) + 12 parameter words + ByteCount(2) + Pad(1).
//
// The server locates the data by walking the data block rather than by trusting
// this field, so a chained write does not depend on it — but a request that
// misdescribed itself would still be a wrong request, so the tests send the right
// value.
const writeAndxRequestDataOffset = header.SMB_HEADER_SIZE + 1 + 24 + 2 + 1

// sendChain sends a batched request carrying every command given, in order, and
// returns the raw reply.
//
// The reply is returned raw rather than decoded because an error response in a
// chain is a WordCount 0 block, which the response structure for that command code
// cannot decode. A client reads the header status before parsing a body for
// exactly that reason, so a test asserting the error case has to work the same way.
func sendChain(t *testing.T, client *smb1client.Client, chain ...command_interface.CommandInterface) []byte {
	t.Helper()

	if len(chain) == 0 {
		t.Fatal("sendChain needs at least one command")
	}

	request := newRequest(chain[0].GetCommandCode())
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	for _, cmd := range chain {
		request.AddCommand(cmd)
	}
	// AddCommand takes the header code from the first command, which newRequest
	// already set; setting it again keeps the two from drifting apart.
	request.Header.Command = chain[0].GetCommandCode()

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the batched request: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if len(raw) < header.SMB_HEADER_SIZE {
		t.Fatalf("the reply is %d bytes, shorter than an SMB header", len(raw))
	}
	return raw
}

// replyStatus reads the status out of a raw reply's header.
func replyStatus(raw []byte) uint32 {
	return binary.LittleEndian.Uint32(raw[5:9])
}

// chainBlockAt describes the AndX linkage of the command block starting at
// position, which is WordCount(1) then AndXCommand(1) AndXReserved(1)
// AndXOffset(2).
func chainBlockAt(t *testing.T, raw []byte, position int) (wordCount int, follows codes.CommandCode, next int) {
	t.Helper()
	if position+5 > len(raw) {
		t.Fatalf("a command block at %d does not fit in the %d-byte reply", position, len(raw))
	}
	return int(raw[position]),
		codes.CommandCode(raw[position+1]),
		int(binary.LittleEndian.Uint16(raw[position+3 : position+5]))
}

// TestBatchedChainExecutesEveryCommand asserts every command of a batch runs and
// that every answer comes back in one message.
//
// A write batched with a close is the shape a client uses to finish with a file in
// one round trip. Before chains were executed the close was decoded and dropped,
// so the write landed, the handle leaked, and the client was told the batch had
// completed.
func TestBatchedChainExecutesEveryCommand(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("batched.txt", []byte("")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("batched.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN_IF,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}

	payload := []byte("written as one half of a batch")

	write := commands.NewWriteAndxRequest()
	write.FID = types.USHORT(fid)
	write.Offset = types.ULONG(0)
	write.DataLength = types.USHORT(len(payload))
	write.DataOffset = types.USHORT(writeAndxRequestDataOffset)
	write.Data = []types.UCHAR(payload)

	closeCmd := commands.NewCloseRequest()
	closeCmd.FID = types.USHORT(fid)

	raw := sendChain(t, client, write, closeCmd)

	if status := replyStatus(raw); status != 0 {
		t.Fatalf("the batch reported 0x%08X, want success", status)
	}
	if got := codes.CommandCode(raw[4]); got != codes.SMB_COM_WRITE_ANDX {
		t.Errorf("the reply header names 0x%02X, want SMB_COM_WRITE_ANDX", uint8(got))
	}

	// The write response links to a close response.
	_, follows, next := chainBlockAt(t, raw, header.SMB_HEADER_SIZE)
	if follows != codes.SMB_COM_CLOSE {
		t.Fatalf("the write response names 0x%02X as its follow-on, want SMB_COM_CLOSE", uint8(follows))
	}
	if next <= header.SMB_HEADER_SIZE || next >= len(raw) {
		t.Fatalf("the close response is at %d, outside the %d-byte reply", next, len(raw))
	}
	// A close response is WordCount 0, ByteCount 0.
	if wordCount := int(raw[next]); wordCount != 0 {
		t.Errorf("the close response has WordCount %d, want 0", wordCount)
	}

	// The close actually happened, so the handle is gone. This is checked before
	// reopening the file: the identifier allocator reuses a released FID, so a
	// reopen can hand back the same number and make a stale handle look live.
	if err := client.CloseFile(fid); err == nil {
		t.Error("closing the batched handle a second time succeeded, so the batched close was dropped")
	}

	// And the write actually happened.
	reopened, err := client.OpenFile("batched.txt",
		fileflags.GENERIC_READ,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("reopening the file failed: %v", err)
	}
	defer client.CloseFile(reopened)

	contents, err := client.ReadFile(reopened, 0, uint32(len(payload)))
	if err != nil {
		t.Fatalf("reading the file back failed: %v", err)
	}
	if !bytes.Equal(contents, payload) {
		t.Errorf("the file holds %q, want %q", contents, payload)
	}
}

// TestBatchedChainStopsAtTheFirstFailure asserts a failing command ends the chain
// and that the answers produced before it are still returned.
//
// [MS-CIFS] section 3.3.4.1: "processing of the AndX request chain terminates with
// the request that generated the error. The error response MUST be the last
// response in the returned AndX chain."
func TestBatchedChainStopsAtTheFirstFailure(t *testing.T) {
	payload := []byte("readable")

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("readable.txt", payload); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("readable.txt",
		fileflags.GENERIC_READ,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}
	defer client.CloseFile(fid)

	good := commands.NewReadAndxRequest()
	good.FID = types.USHORT(fid)
	good.MaxCountOfBytesToReturn = types.USHORT(len(payload))

	// A FID the connection never handed out.
	bad := commands.NewReadAndxRequest()
	bad.FID = types.USHORT(0xBEEF)
	bad.MaxCountOfBytesToReturn = types.USHORT(16)

	third := commands.NewReadAndxRequest()
	third.FID = types.USHORT(fid)
	third.MaxCountOfBytesToReturn = types.USHORT(len(payload))

	raw := sendChain(t, client, good, bad, third)

	if status := replyStatus(raw); status != uint32(nt_status.NT_STATUS_SMB_BAD_FID) {
		t.Fatalf("the batch reported 0x%08X, want STATUS_SMB_BAD_FID (0x%08X)",
			status, uint32(nt_status.NT_STATUS_SMB_BAD_FID))
	}

	// The first read answered, and links to the error response.
	wordCount, follows, next := chainBlockAt(t, raw, header.SMB_HEADER_SIZE)
	if wordCount == 0 {
		t.Fatal("the first read produced an error response, so the chain never got to the failing command")
	}
	if follows != codes.SMB_COM_READ_ANDX {
		t.Fatalf("the first response names 0x%02X as its follow-on, want SMB_COM_READ_ANDX", uint8(follows))
	}

	// The error response is a WordCount 0, ByteCount 0 block, and it is the last
	// thing in the message: the third command must not have run.
	if next+3 > len(raw) {
		t.Fatalf("the error response at %d does not fit in the %d-byte reply", next, len(raw))
	}
	if got := raw[next : next+3]; !bytes.Equal(got, []byte{0x00, 0x00, 0x00}) {
		t.Fatalf("the block at %d is % x, want an error response of 00 00 00", next, got)
	}
	if next+3 != len(raw) {
		t.Fatalf("the reply has %d bytes after the error response, want none — a command after the failure ran",
			len(raw)-(next+3))
	}
}

// TestBatchedChainSeesIdentifiersAssignedWithinIt asserts a command batched behind
// a tree connect acts on the tree the tree connect just created.
//
// The client cannot know the TID when it builds the batch — the server has not
// assigned one yet — so it sends whatever it has and the server has to substitute
// what it allocated. SMB_COM_CHECK_DIRECTORY is one of the commands [MS-CIFS]
// section 2.2.4.55 documents as following a tree connect.
func TestBatchedChainSeesIdentifiersAssignedWithinIt(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddDirectory("adirectory"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	connect := commands.NewTreeConnectAndxRequest()
	connect.Password = []types.UCHAR{}
	connect.PasswordLength = types.USHORT(0)
	connect.Path = []types.UCHAR("\\\\127.0.0.1\\" + fileShareName + "\x00")
	connect.Service = []types.UCHAR("?????\x00")

	check := commands.NewCheckDirectoryRequest()
	if err := check.DirectoryName.SetString("adirectory"); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	// A TID the server never issued, so an answer of success can only come from
	// the tree the batch itself connected.
	request := newRequest(codes.SMB_COM_TREE_CONNECT_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = 0xFFFE
	// Both names above are ASCII, and a name is read in the encoding the message
	// declares, so the message must not declare Unicode.
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)
	request.AddCommand(connect)
	request.AddCommand(check)
	request.Header.Command = codes.SMB_COM_TREE_CONNECT_ANDX

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the batched request: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}

	if status := replyStatus(raw); status != 0 {
		t.Fatalf("the batch reported 0x%08X, want success — the chained command did not see the new tree", status)
	}

	_, follows, next := chainBlockAt(t, raw, header.SMB_HEADER_SIZE)
	if follows != codes.SMB_COM_CHECK_DIRECTORY {
		t.Fatalf("the tree connect response names 0x%02X as its follow-on, want SMB_COM_CHECK_DIRECTORY",
			uint8(follows))
	}
	if next == 0 || next >= len(raw) {
		t.Fatalf("the chained response is at %d, outside the %d-byte reply", next, len(raw))
	}

	// The reply's TID is the one the server allocated, and it is not the invented
	// one the request carried.
	if tid := binary.LittleEndian.Uint16(raw[28:30]); tid == 0xFFFE {
		t.Errorf("the reply echoes the invented TID 0x%04X rather than the one it allocated", tid)
	}
}

// TestUnbatchedAndXRequestTerminatesItsChain asserts a single AndX command still
// answers with a terminated one-command chain, so a request that was never batched
// is unaffected by chain support.
func TestUnbatchedAndXRequestTerminatesItsChain(t *testing.T) {
	payload := []byte("unbatched")

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("single.txt", payload); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("single.txt",
		fileflags.GENERIC_READ,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}
	defer client.CloseFile(fid)

	read := commands.NewReadAndxRequest()
	read.FID = types.USHORT(fid)
	read.MaxCountOfBytesToReturn = types.USHORT(len(payload))

	raw := sendChain(t, client, read)

	if status := replyStatus(raw); status != 0 {
		t.Fatalf("the read reported 0x%08X, want success", status)
	}

	_, follows, next := chainBlockAt(t, raw, header.SMB_HEADER_SIZE)
	if follows != codes.SMB_COM_NO_ANDX_COMMAND {
		t.Errorf("a lone response names 0x%02X as its follow-on, want SMB_COM_NO_ANDX_COMMAND", uint8(follows))
	}
	if next != 0 {
		t.Errorf("a lone response has AndXOffset %d, want 0", next)
	}

	// And the data is where the response says it is, which is the property a
	// chained response has to preserve too.
	decoded := message.NewMessage()
	if err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	response, ok := decoded.Command.(*commands.ReadAndxResponse)
	if !ok {
		t.Fatalf("the reply carries %T, want *ReadAndxResponse", decoded.Command)
	}
	at := int(response.DataOffset)
	if at+len(payload) > len(raw) {
		t.Fatalf("DataOffset %d plus %d bytes runs past the %d-byte reply", at, len(payload), len(raw))
	}
	if got := raw[at : at+len(payload)]; !bytes.Equal(got, payload) {
		t.Errorf("DataOffset %d points at %q, want %q", at, got, payload)
	}
}
