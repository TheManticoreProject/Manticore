package server

import (
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// coreFileServer serves one file of known contents.
func coreFileServer(t *testing.T, readOnly bool) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("adirectory"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, readOnly)
	return fs, client
}

// coreOpenFile opens a file with SMB_COM_OPEN and returns its FID.
func coreOpenFile(t *testing.T, client *smb1client.Client, name string, accessMode uint16) (uint16, uint32) {
	t.Helper()

	request := commands.NewOpenRequest()
	request.AccessMode = types.USHORT(accessMode)
	if err := request.FileName.SetString(name); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	status, body := sendLegacy(t, client, codes.SMB_COM_OPEN, request)
	if status != 0 {
		return 0, status
	}

	response := commands.NewOpenResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the open reply did not decode: %v", err)
	}
	return uint16(response.FID), 0
}

// TestCoreOpenReadWriteRoundTrip asserts the core-set open, write and read work
// together on one handle.
func TestCoreOpenReadWriteRoundTrip(t *testing.T) {
	fs, client := coreFileServer(t, false)

	fid, status := coreOpenFile(t, client, "existing.txt", openAccessReadWrite)
	if status != 0 {
		t.Fatalf("SMB_COM_OPEN reported 0x%08X", status)
	}

	payload := []byte("core-set write")

	write := commands.NewWriteRequest()
	write.FID = types.USHORT(fid)
	write.CountOfBytesToWrite = types.USHORT(len(payload))
	write.WriteOffsetInBytes = types.ULONG(0)
	write.Data.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	write.Data.Buffer = []types.UCHAR(payload)

	status, body := sendLegacy(t, client, codes.SMB_COM_WRITE, write)
	if status != 0 {
		t.Fatalf("SMB_COM_WRITE reported 0x%08X", status)
	}
	writeResponse := commands.NewWriteResponse()
	if _, err := writeResponse.Unmarshal(body); err != nil {
		t.Fatalf("the write reply did not decode: %v", err)
	}
	if int(writeResponse.CountOfBytesWritten) != len(payload) {
		t.Errorf("the reply reports %d bytes written, want %d",
			writeResponse.CountOfBytesWritten, len(payload))
	}

	read := commands.NewReadRequest()
	read.FID = types.SHORT(fid)
	read.CountOfBytesToRead = types.USHORT(len(payload))
	read.ReadOffsetInBytes = types.ULONG(0)

	status, body = sendLegacy(t, client, codes.SMB_COM_READ, read)
	if status != 0 {
		t.Fatalf("SMB_COM_READ reported 0x%08X", status)
	}
	readResponse := commands.NewReadResponse()
	if _, err := readResponse.Unmarshal(body); err != nil {
		t.Fatalf("the read reply did not decode: %v", err)
	}
	if int(readResponse.CountOfBytesReturned) != len(payload) {
		t.Fatalf("the read returned %d bytes, want %d",
			readResponse.CountOfBytesReturned, len(payload))
	}
	if got := string(readResponse.Bytes.Buffer); got != string(payload) {
		t.Errorf("the read returned %q, want %q", got, payload)
	}

	// And the bytes landed in the backend, not just in the reply.
	stored, err := fs.Open("existing.txt", OpenFlags{Read: true})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer stored.Close()

	buffer := make([]byte, len(payload))
	if _, err := stored.ReadAt(buffer, 0); err != nil {
		t.Fatalf("ReadAt() error = %v", err)
	}
	if string(buffer) != string(payload) {
		t.Errorf("the file holds %q, want %q", buffer, payload)
	}
}

// TestCoreOpenRefusesAReservedAccessMode asserts a reserved access value is
// refused rather than treated as read.
//
// [MS-CIFS] section 2.2.4.3.1 requires values 4 to 7 to be answered with
// STATUS_OS2_INVALID_ACCESS. Defaulting one of them to read would grant an access
// the client neither asked for nor expects.
func TestCoreOpenRefusesAReservedAccessMode(t *testing.T) {
	_, client := coreFileServer(t, false)

	for _, mode := range []uint16{4, 5, 6, 7} {
		_, status := coreOpenFile(t, client, "existing.txt", mode)
		if status != uint32(nt_status.NT_STATUS_OS2_INVALID_ACCESS) {
			t.Errorf("access mode %d reported 0x%08X, want STATUS_OS2_INVALID_ACCESS (0x%08X)",
				mode, status, uint32(nt_status.NT_STATUS_OS2_INVALID_ACCESS))
		}
	}
}

// TestCoreWriteOfZeroBytesTruncates asserts a zero-length write sets the file's
// length rather than doing nothing.
//
// [MS-CIFS] section 3.3.5.14 defines a CountOfBytesToWrite of zero as a request to
// set the length, so treating it as a no-op silently drops a truncation the client
// believes succeeded.
func TestCoreWriteOfZeroBytesTruncates(t *testing.T) {
	fs, client := coreFileServer(t, false)

	fid, status := coreOpenFile(t, client, "existing.txt", openAccessReadWrite)
	if status != 0 {
		t.Fatalf("SMB_COM_OPEN reported 0x%08X", status)
	}

	write := commands.NewWriteRequest()
	write.FID = types.USHORT(fid)
	write.CountOfBytesToWrite = types.USHORT(0)
	write.WriteOffsetInBytes = types.ULONG(4)
	write.Data.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)

	if status, _ := sendLegacy(t, client, codes.SMB_COM_WRITE, write); status != 0 {
		t.Fatalf("SMB_COM_WRITE reported 0x%08X", status)
	}

	attr, err := fs.Stat("existing.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if attr.Size != 4 {
		t.Errorf("the file is %d bytes after a zero-length write at 4, want 4", attr.Size)
	}
}

// TestCoreCreateAndCreateNewDifferOnExistence asserts the two create commands
// differ in exactly the way their names say.
func TestCoreCreateAndCreateNewDifferOnExistence(t *testing.T) {
	fs, client := coreFileServer(t, false)

	t.Run("create truncates an existing file", func(t *testing.T) {
		request := commands.NewCreateRequest()
		if err := request.FileName.SetString("existing.txt"); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		status, body := sendLegacy(t, client, codes.SMB_COM_CREATE, request)
		if status != 0 {
			t.Fatalf("SMB_COM_CREATE reported 0x%08X", status)
		}
		response := commands.NewCreateResponse()
		if _, err := response.Unmarshal(body); err != nil {
			t.Fatalf("the reply did not decode: %v", err)
		}

		attr, err := fs.Stat("existing.txt")
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if attr.Size != 0 {
			t.Errorf("the file is %d bytes after a create, want it truncated", attr.Size)
		}
	})

	t.Run("create new refuses an existing file", func(t *testing.T) {
		request := commands.NewCreateNewRequest()
		if err := request.FileName.SetString("existing.txt"); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		if status, _ := sendLegacy(t, client, codes.SMB_COM_CREATE_NEW, request); status == 0 {
			t.Error("SMB_COM_CREATE_NEW succeeded on a file that already exists")
		}
	})

	t.Run("create new makes a missing file", func(t *testing.T) {
		request := commands.NewCreateNewRequest()
		if err := request.FileName.SetString("brandnew.txt"); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		if status, _ := sendLegacy(t, client, codes.SMB_COM_CREATE_NEW, request); status != 0 {
			t.Fatalf("SMB_COM_CREATE_NEW reported 0x%08X", status)
		}
		if _, err := fs.Stat("brandnew.txt"); err != nil {
			t.Errorf("the file was not created: %v", err)
		}
	})
}

// TestCoreOpenAndxHonoursItsOpenMode asserts the two OpenMode fields decide
// existence handling, which is the create-disposition of the NT commands in two
// bits.
func TestCoreOpenAndxHonoursItsOpenMode(t *testing.T) {
	fs, client := coreFileServer(t, false)

	openAndx := func(t *testing.T, name string, openMode uint16) uint32 {
		t.Helper()

		request := commands.NewOpenAndxRequest()
		request.AccessMode = types.USHORT(openAccessReadWrite)
		request.OpenMode = types.USHORT(openMode)
		// The name is a null-terminated ASCII string ([MS-CIFS] section
		// 2.2.4.41.1). This command leaves the format to its caller rather than
		// setting it in Marshal as most do, so a client has to say so.
		request.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
		if err := request.FileName.SetString(name); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}
		status, _ := sendLegacy(t, client, codes.SMB_COM_OPEN_ANDX, request)
		return status
	}

	t.Run("truncate", func(t *testing.T) {
		if status := openAndx(t, "existing.txt", openModeTruncate); status != 0 {
			t.Fatalf("a truncating open reported 0x%08X", status)
		}
		attr, err := fs.Stat("existing.txt")
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if attr.Size != 0 {
			t.Errorf("the file is %d bytes after a truncating open, want 0", attr.Size)
		}
	})

	t.Run("create if missing", func(t *testing.T) {
		if status := openAndx(t, "made-by-openandx.txt", openModeCreateIfMissing); status != 0 {
			t.Fatalf("a creating open reported 0x%08X", status)
		}
		if _, err := fs.Stat("made-by-openandx.txt"); err != nil {
			t.Errorf("the file was not created: %v", err)
		}
	})

	t.Run("refuse a missing file when not asked to create", func(t *testing.T) {
		if status := openAndx(t, "notthere.txt", 0); status == 0 {
			t.Error("an open of a missing file succeeded without the create bit")
		}
	})
}

// TestCoreWriteAndCloseReleasesTheHandle asserts the handle is gone afterwards,
// which is the command's whole contract.
func TestCoreWriteAndCloseReleasesTheHandle(t *testing.T) {
	fs, client := coreFileServer(t, false)

	fid, status := coreOpenFile(t, client, "existing.txt", openAccessReadWrite)
	if status != 0 {
		t.Fatalf("SMB_COM_OPEN reported 0x%08X", status)
	}

	payload := []byte("written and closed")

	request := commands.NewWriteAndCloseRequest()
	request.FID = types.USHORT(fid)
	request.CountOfBytesToWrite = types.USHORT(len(payload))
	request.WriteOffsetInBytes = types.ULONG(0)
	request.Data = []types.UCHAR(payload)

	status, body := sendLegacy(t, client, codes.SMB_COM_WRITE_AND_CLOSE, request)
	if status != 0 {
		t.Fatalf("SMB_COM_WRITE_AND_CLOSE reported 0x%08X", status)
	}
	response := commands.NewWriteAndCloseResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	if int(response.CountOfBytesWritten) != len(payload) {
		t.Errorf("the reply reports %d bytes written, want %d",
			response.CountOfBytesWritten, len(payload))
	}

	// The handle is gone: a read through it now fails.
	read := commands.NewReadRequest()
	read.FID = types.SHORT(fid)
	read.CountOfBytesToRead = types.USHORT(4)
	if status, _ := sendLegacy(t, client, codes.SMB_COM_READ, read); status == 0 {
		t.Error("reading through the closed handle succeeded")
	}

	attr, err := fs.Stat("existing.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if int(attr.Size) != len(payload) {
		t.Errorf("the file is %d bytes, want %d", attr.Size, len(payload))
	}
}

// TestCoreSeekMovesAndRemembersThePosition asserts each seek mode, and that the
// position persists between calls.
//
// Seeking from the current position is only meaningful if the position is
// remembered, so a server that computed an answer without storing it would give
// different results for the same sequence of requests.
func TestCoreSeekMovesAndRemembersThePosition(t *testing.T) {
	_, client := coreFileServer(t, false)

	fid, status := coreOpenFile(t, client, "existing.txt", openAccessRead)
	if status != 0 {
		t.Fatalf("SMB_COM_OPEN reported 0x%08X", status)
	}

	seek := func(t *testing.T, mode uint16, offset int32) (uint32, uint32) {
		t.Helper()

		request := commands.NewSeekRequest()
		request.FID = types.USHORT(fid)
		request.Mode = types.USHORT(mode)
		request.Offset = types.LONG(offset)

		status, body := sendLegacy(t, client, codes.SMB_COM_SEEK, request)
		if status != 0 {
			return 0, status
		}
		response := commands.NewSeekResponse()
		if _, err := response.Unmarshal(body); err != nil {
			t.Fatalf("the seek reply did not decode: %v", err)
		}
		return uint32(response.Offset), 0
	}

	if got, status := seek(t, seekFromStart, 4); status != 0 || got != 4 {
		t.Fatalf("seeking to 4 from the start gave %d (status 0x%08X), want 4", got, status)
	}
	// From the current position, which is only right if the previous seek stuck.
	if got, status := seek(t, seekFromCurrent, 3); status != 0 || got != 7 {
		t.Fatalf("seeking 3 from the current position gave %d (status 0x%08X), want 7", got, status)
	}
	// The file is ten bytes, so the end minus two is eight.
	if got, status := seek(t, seekFromEnd, -2); status != 0 || got != 8 {
		t.Fatalf("seeking -2 from the end gave %d (status 0x%08X), want 8", got, status)
	}

	// Before the start is refused rather than clamped: a clamped answer would
	// report a seek the client did not ask for as a success.
	if _, status := seek(t, seekFromStart, -1); status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
		t.Errorf("seeking before the start reported 0x%08X, want STATUS_INVALID_PARAMETER", status)
	}
}

// TestCoreTreeConnectConnectsSameAsTheAndXForm asserts the deprecated tree connect
// returns a usable TID.
func TestCoreTreeConnectConnectsSameAsTheAndXForm(t *testing.T) {
	_, client := coreFileServer(t, false)

	// All three strings are null-terminated ASCII ([MS-CIFS] section 2.2.4.50.1),
	// and this command leaves the format to its caller.
	request := commands.NewTreeConnectRequest()
	request.Path.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	request.Path.SetString(`\\127.0.0.1\` + fileShareName)
	request.Password.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	request.Password.SetString("")
	request.Service.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	request.Service.SetString("?????")

	status, body := sendLegacy(t, client, codes.SMB_COM_TREE_CONNECT, request)
	if status != 0 {
		t.Fatalf("SMB_COM_TREE_CONNECT reported 0x%08X", status)
	}

	response := commands.NewTreeConnectResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	if response.TID == 0 {
		t.Error("the reply carries no TID")
	}
	if response.MaxBufferSize == 0 {
		t.Error("the reply carries no MaxBufferSize")
	}
}

// TestCoreProcessExitReleasesThatProcessesHandles asserts the command closes what
// the request's PID held, and leaves another process's handles alone.
//
// [MS-CIFS] section 2.2.4.18 has the server close "any resources owned by the
// Process ID (PID) listed in the request header". Closing everything, or nothing,
// would both be wrong: the command exists to clean up after one failed process on
// a connection that may be serving several.
func TestCoreProcessExitReleasesThatProcessesHandles(t *testing.T) {
	_, client := coreFileServer(t, false)

	// newRequest uses one PID for every request, so the handle below belongs to
	// it and the exit that follows names the same one.
	fid, status := coreOpenFile(t, client, "existing.txt", openAccessRead)
	if status != 0 {
		t.Fatalf("SMB_COM_OPEN reported 0x%08X", status)
	}

	if status, _ := sendLegacy(t, client, codes.SMB_COM_PROCESS_EXIT, commands.NewProcessExitRequest()); status != 0 {
		t.Fatalf("SMB_COM_PROCESS_EXIT reported 0x%08X", status)
	}

	read := commands.NewReadRequest()
	read.FID = types.SHORT(fid)
	read.CountOfBytesToRead = types.USHORT(4)
	if status, _ := sendLegacy(t, client, codes.SMB_COM_READ, read); status == 0 {
		t.Error("the handle survived the exit of the process that opened it")
	}

	// The connection is still usable afterwards.
	if _, err := client.Echo([]byte("alive")); err != nil {
		t.Fatalf("the connection did not survive a process exit: %v", err)
	}
}

// TestCoreOpenOnAReadOnlyShareRefusesWriting asserts the share's refusal applies to
// the core-set opens too.
func TestCoreOpenOnAReadOnlyShareRefusesWriting(t *testing.T) {
	_, client := coreFileServer(t, true)

	if _, status := coreOpenFile(t, client, "existing.txt", openAccessWrite); status == 0 {
		t.Error("a write open succeeded on a read-only share")
	}
	if _, status := coreOpenFile(t, client, "existing.txt", openAccessRead); status != 0 {
		t.Errorf("a read open on a read-only share reported 0x%08X", status)
	}
}
