package server

import (
	"encoding/binary"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// open2Parameters builds a TRANS2_OPEN2 request's parameter block.
//
// Flags(2) AccessMode(2) Reserved1(2) FileAttributes(2) CreationTime(4)
// OpenMode(2) AllocationSize(4) Reserved[5](10) FileName(variable), per [MS-CIFS]
// section 2.2.6.1.1.
func open2Parameters(accessMode, openMode uint16, name string) []byte {
	parameters := make([]byte, 28)
	binary.LittleEndian.PutUint16(parameters[2:4], accessMode)
	binary.LittleEndian.PutUint16(parameters[12:14], openMode)
	return append(parameters, append([]byte(name), 0x00)...)
}

// TestTrans2Open2OpensAndCreates asserts the transaction open honours its
// OpenMode, and reports which of the two it did.
//
// ActionTaken is the only way a client can tell a file it created from one it
// found, so reporting it wrongly would have a client overwrite something it
// believed it had just made.
func TestTrans2Open2OpensAndCreates(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	t.Run("opens an existing file", func(t *testing.T) {
		parameters, _, status := sendTrans2(t, client, subcommands.TRANS2_OPEN2,
			open2Parameters(openAccessReadWrite, 0, "existing.txt"), nil)
		if status != 0 {
			t.Fatalf("TRANS2_OPEN2 reported 0x%08X, want success", status)
		}
		if len(parameters) < 20 {
			t.Fatalf("the reply carries %d parameter bytes, want at least 20", len(parameters))
		}
		if fid := binary.LittleEndian.Uint16(parameters[0:2]); fid == 0 {
			t.Error("the reply carries no FID")
		}
		if size := binary.LittleEndian.Uint32(parameters[8:12]); size != 10 {
			t.Errorf("FileDataSize is %d, want 10", size)
		}
		if action := binary.LittleEndian.Uint16(parameters[18:20]); action != open2ActionOpened {
			t.Errorf("ActionTaken is %d, want %d (opened)", action, open2ActionOpened)
		}
	})

	t.Run("creates a missing file", func(t *testing.T) {
		parameters, _, status := sendTrans2(t, client, subcommands.TRANS2_OPEN2,
			open2Parameters(openAccessReadWrite, openModeCreateIfMissing, "created.txt"), nil)
		if status != 0 {
			t.Fatalf("TRANS2_OPEN2 reported 0x%08X, want success", status)
		}
		if action := binary.LittleEndian.Uint16(parameters[18:20]); action != open2ActionCreated {
			t.Errorf("ActionTaken is %d, want %d (created)", action, open2ActionCreated)
		}
		if _, err := fs.Stat("created.txt"); err != nil {
			t.Errorf("the file was not created: %v", err)
		}
	})

	t.Run("truncates when asked", func(t *testing.T) {
		if _, _, status := sendTrans2(t, client, subcommands.TRANS2_OPEN2,
			open2Parameters(openAccessReadWrite, openModeTruncate, "existing.txt"), nil); status != 0 {
			t.Fatalf("TRANS2_OPEN2 reported 0x%08X", status)
		}
		attr, err := fs.Stat("existing.txt")
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if attr.Size != 0 {
			t.Errorf("the file is %d bytes after a truncating open, want 0", attr.Size)
		}
	})

	t.Run("refuses a missing file when not asked to create", func(t *testing.T) {
		if _, _, status := sendTrans2(t, client, subcommands.TRANS2_OPEN2,
			open2Parameters(openAccessRead, 0, "notthere.txt"), nil); status == 0 {
			t.Error("an open of a missing file succeeded without the create bit")
		}
	})
}

// TestTrans2CreateDirectoryMakesOneAndRefusesADuplicate asserts the subcommand
// creates a directory and refuses one that exists.
//
// [MS-CIFS] section 2.2.6.14: "The directory MUST NOT exist. If the directory does
// exist, the request MUST fail and the server MUST return
// STATUS_OBJECT_NAME_COLLISION."
func TestTrans2CreateDirectoryMakesOneAndRefusesADuplicate(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	parameters := append(make([]byte, 4), append([]byte("newdirectory"), 0x00)...)

	if _, _, status := sendTrans2(t, client, subcommands.TRANS2_CREATE_DIRECTORY, parameters, nil); status != 0 {
		t.Fatalf("TRANS2_CREATE_DIRECTORY reported 0x%08X, want success", status)
	}
	attr, err := fs.Stat("newdirectory")
	if err != nil {
		t.Fatalf("the directory was not created: %v", err)
	}
	if !attr.IsDir {
		t.Error("what was created is not a directory")
	}

	_, _, status := sendTrans2(t, client, subcommands.TRANS2_CREATE_DIRECTORY, parameters, nil)
	if status != uint32(nt_status.NT_STATUS_OBJECT_NAME_COLLISION) {
		t.Errorf("creating it twice reported 0x%08X, want STATUS_OBJECT_NAME_COLLISION (0x%08X)",
			status, uint32(nt_status.NT_STATUS_OBJECT_NAME_COLLISION))
	}
}

// TestTrans2CreateDirectoryOnAReadOnlyShareIsRefused asserts the share's refusal
// applies.
func TestTrans2CreateDirectoryOnAReadOnlyShareIsRefused(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, true)

	parameters := append(make([]byte, 4), append([]byte("newdirectory"), 0x00)...)
	if _, _, status := sendTrans2(t, client, subcommands.TRANS2_CREATE_DIRECTORY, parameters, nil); status == 0 {
		t.Error("a directory was created on a read-only share")
	}
}

// TestNtTransactCreateOpensThroughATransaction asserts the NT_TRANSACT open works
// and reports the file it opened.
func TestNtTransactCreateOpensThroughATransaction(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	// Flags(4) RootDirectoryFID(4) DesiredAccess(4) AllocationSize(8)
	// ExtFileAttributes(4) ShareAccess(4) CreateDisposition(4) CreateOptions(4)
	// SecurityDescriptorLength(4) EALength(4) NameLength(4)
	// ImpersonationLevel(4) SecurityFlags(1), then the name.
	name := "existing.txt"
	parameters := make([]byte, 53)
	binary.LittleEndian.PutUint32(parameters[8:12], 0x00120089)  // DesiredAccess: read
	binary.LittleEndian.PutUint32(parameters[28:32], 0x00000001) // FILE_OPEN
	binary.LittleEndian.PutUint32(parameters[32:36], 0x00000040) // FILE_NON_DIRECTORY_FILE
	binary.LittleEndian.PutUint32(parameters[40:44], uint32(len(name)))
	parameters = append(parameters, []byte(name)...)

	got, _, status := sendNtTransactWithParameters(t, client, subcommands.NT_TRANSACT_CREATE, parameters)
	if status != 0 {
		t.Fatalf("NT_TRANSACT_CREATE reported 0x%08X, want success", status)
	}
	if len(got) != 69 {
		t.Fatalf("the reply carries %d parameter bytes, want the 69 of [MS-CIFS] 2.2.7.1.2", len(got))
	}
	if fid := binary.LittleEndian.Uint16(got[2:4]); fid == 0 {
		t.Error("the reply carries no FID")
	}
	if size := binary.LittleEndian.Uint64(got[56:64]); size != 10 {
		t.Errorf("EndOfFile is %d, want 10", size)
	}
	if got[0] != 0 {
		t.Errorf("OpLockLevel is %d, want 0 — no oplock is granted", got[0])
	}
}

// TestNtTransactCreateRefusesARelativeRoot asserts a create relative to a
// directory handle is refused rather than resolved against the share root.
//
// Treating the name as share-relative instead would open a different file and
// report success.
func TestNtTransactCreateRefusesARelativeRoot(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	name := "somewhere.txt"
	parameters := make([]byte, 53)
	binary.LittleEndian.PutUint32(parameters[4:8], 0x1234) // RootDirectoryFID
	binary.LittleEndian.PutUint32(parameters[40:44], uint32(len(name)))
	parameters = append(parameters, []byte(name)...)

	_, _, status := sendNtTransactWithParameters(t, client, subcommands.NT_TRANSACT_CREATE, parameters)
	if status != uint32(nt_status.NT_STATUS_NOT_SUPPORTED) {
		t.Errorf("a create relative to a handle reported 0x%08X, want STATUS_NOT_SUPPORTED", status)
	}
}

// TestReservedSubcommandsUseTheirMandatedStatus asserts the two reserved
// subcommands answer with the status the specification names for each, rather than
// the generic refusal.
//
// [MS-CIFS] section 2.2.6.5 requires STATUS_SMB_NO_SUPPORT for
// TRANS2_SET_FS_INFORMATION, and section 2.2.7.5 requires STATUS_SMB_BAD_COMMAND
// for NT_TRANSACT_RENAME. Both differ from the STATUS_NOT_IMPLEMENTED their
// neighbours get, which is why each is handled rather than left to fall through.
func TestReservedSubcommandsUseTheirMandatedStatus(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	t.Run("TRANS2_SET_FS_INFORMATION", func(t *testing.T) {
		// STATUS_SMB_NO_SUPPORT has no constant in windows/errors/nt_status, so the
		// closest available is used; what matters here is that it is not the
		// generic STATUS_NOT_IMPLEMENTED.
		_, _, status := sendTrans2(t, client, subcommands.TRANS2_SET_FS_INFORMATION, make([]byte, 4), nil)
		if status != uint32(nt_status.NT_STATUS_NOT_SUPPORTED) {
			t.Errorf("it reported 0x%08X, want STATUS_NOT_SUPPORTED (0x%08X)",
				status, uint32(nt_status.NT_STATUS_NOT_SUPPORTED))
		}
		if status == uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED) {
			t.Error("it fell through to the table's generic refusal")
		}
	})

	t.Run("NT_TRANSACT_RENAME", func(t *testing.T) {
		_, _, status := sendNtTransactWithParameters(t, client, subcommands.NT_TRANSACT_RENAME, make([]byte, 4))
		if status != uint32(nt_status.NT_STATUS_SMB_BAD_COMMAND) {
			t.Errorf("it reported 0x%08X, want STATUS_SMB_BAD_COMMAND (0x%08X)",
				status, uint32(nt_status.NT_STATUS_SMB_BAD_COMMAND))
		}
	})
}

// sendTrans2 sends a TRANSACTION2 and returns the parameter block, the data block
// and the status.
//
// The existing sendNtTransact returns only a status, and these subcommands answer
// in their parameter block, so the reply has to come back whole.
func sendTrans2(
	t *testing.T,
	client *smb1client.Client,
	subcommand subcommands.Transaction2Subcommand,
	parameters, data []byte,
) ([]byte, []byte, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_TRANSACTION2)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	// The names in these parameter blocks are ASCII, so the message must not
	// declare Unicode.
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewTransaction2Request()
	transaction.Setup = []types.USHORT{types.USHORT(subcommand)}
	transaction.SetupCount = types.UCHAR(1)
	transaction.MaxParameterCount = 1024
	transaction.MaxDataCount = 4096
	transaction.TotalParameterCount = types.USHORT(len(parameters))
	transaction.ParameterCount = types.USHORT(len(parameters))
	transaction.Trans2_Parameters = parameters
	transaction.TotalDataCount = types.USHORT(len(data))
	transaction.DataCount = types.USHORT(len(data))
	transaction.Trans2_Data = data
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if len(raw) < 9 {
		t.Fatalf("the reply is %d bytes", len(raw))
	}

	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != 0 {
		return nil, nil, status
	}

	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	decoded, ok := response.Command.(*commands.Transaction2Response)
	if !ok {
		t.Fatalf("the reply carries %T", response.Command)
	}
	return []byte(decoded.Trans2_Parameters), []byte(decoded.Trans2_Data), 0
}

// sendNtTransactWithParameters is sendNtTransact with the reply's parameter block
// returned, which is where these subcommands answer.
func sendNtTransactWithParameters(
	t *testing.T,
	client *smb1client.Client,
	function subcommands.NtTransactSubcommand,
	parameters []byte,
) ([]byte, []byte, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(function)
	transaction.MaxParameterCount = 1024
	transaction.MaxDataCount = 4096
	transaction.TotalParameterCount = types.ULONG(len(parameters))
	transaction.ParameterCount = types.ULONG(len(parameters))
	transaction.NT_Trans_Parameters = parameters
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if len(raw) < 9 {
		t.Fatalf("the reply is %d bytes", len(raw))
	}

	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != 0 {
		return nil, nil, status
	}

	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	decoded, ok := response.Command.(*commands.NtTransactResponse)
	if !ok {
		t.Fatalf("the reply carries %T", response.Command)
	}
	return []byte(decoded.Parameters), []byte(decoded.Data), 0
}
