package server

import (
	"encoding/binary"
	"testing"
	"time"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// sendLegacy sends a hand-built request and returns the reply's status and body.
//
// These commands have no client helper — the client speaks TRANSACTION2 for
// information — so the tests build the requests directly. The names are sent OEM,
// so the message must not declare Unicode.
func sendLegacy(
	t *testing.T,
	client *smb1client.Client,
	code codes.CommandCode,
	cmd command_interface.CommandInterface,
) (uint32, []byte) {
	t.Helper()

	request := newRequest(code)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal command 0x%02X: %v", uint8(code), err)
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
	return binary.LittleEndian.Uint32(raw[5:9]), raw[header.SMB_HEADER_SIZE:]
}

// legacyInfoServer serves one file with known contents and timestamps.
func legacyInfoServer(t *testing.T, readOnly bool) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("described.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("adirectory"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, readOnly)
	return fs, client
}

// pathRequest builds a request naming a file by path.
func pathRequest(t *testing.T, cmd interface{ SetString(string) error }, path string) {
	t.Helper()
	if err := cmd.SetString(path); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
}

// TestQueryInformationReportsAttributesWriteTimeAndSize asserts the core-set query
// answers with what the backend reports.
func TestQueryInformationReportsAttributesWriteTimeAndSize(t *testing.T) {
	fs, client := legacyInfoServer(t, false)

	request := commands.NewQueryInformationRequest()
	pathRequest(t, &request.FileName, "described.txt")

	status, body := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION, request)
	if status != 0 {
		t.Fatalf("the query reported 0x%08X, want success", status)
	}

	response := commands.NewQueryInformationResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}

	attr, err := fs.Stat("described.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if got := int(response.FileSize); got != int(attr.Size) {
		t.Errorf("FileSize is %d, want %d", got, attr.Size)
	}
	if got := response.FileAttributes.GetAttributes(); got&smbFileAttributeDirectory != 0 {
		t.Errorf("a file reports attributes 0x%04X, which claims a directory", got)
	}
	// UTIME is seconds since 1970, and the backend's stamp is not the zero time.
	if response.LastWriteTime == 0 {
		t.Error("LastWriteTime is zero for a file with a modification time")
	}
	if want := utimeOf(attr.Modified); uint32(response.LastWriteTime) != want {
		t.Errorf("LastWriteTime is %d, want %d", uint32(response.LastWriteTime), want)
	}
}

// TestQueryInformationReportsADirectory asserts a directory is described as one,
// in the 16-bit attribute field rather than the 32-bit one.
//
// SMB_FILE_ATTRIBUTE_NORMAL is 0x0000 in this field, not the extended form's
// 0x0080, so reporting the extended value would set a bit [MS-CIFS] section
// 2.2.1.2.4 reserves.
func TestQueryInformationReportsADirectory(t *testing.T) {
	_, client := legacyInfoServer(t, false)

	request := commands.NewQueryInformationRequest()
	pathRequest(t, &request.FileName, "adirectory")

	status, body := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION, request)
	if status != 0 {
		t.Fatalf("the query reported 0x%08X, want success", status)
	}

	response := commands.NewQueryInformationResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}

	attributes := response.FileAttributes.GetAttributes()
	if attributes&smbFileAttributeDirectory == 0 {
		t.Errorf("a directory reports attributes 0x%04X, which does not say directory", attributes)
	}
	if attributes&^uint16(smbFileAttributeDirectory|smbFileAttributeReadOnly) != 0 {
		t.Errorf("attributes 0x%04X set a bit outside SMB_FILE_ATTRIBUTES", attributes)
	}
}

// TestQueryInformationOnAMissingFileIsRefused asserts a path that does not exist
// is reported rather than answered with zeroes.
func TestQueryInformationOnAMissingFileIsRefused(t *testing.T) {
	_, client := legacyInfoServer(t, false)

	request := commands.NewQueryInformationRequest()
	pathRequest(t, &request.FileName, "nosuchfile.txt")

	status, _ := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION, request)
	if status != uint32(nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND) {
		t.Errorf("querying a missing file reported 0x%08X, want STATUS_OBJECT_NAME_NOT_FOUND (0x%08X)",
			status, uint32(nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND))
	}
}

// TestSetInformationAppliesAttributesAndWriteTime asserts the core-set change is
// applied, and that a zero write time leaves the stamp alone.
//
// [MS-CIFS] section 3.3.5.12: "If this field contains 0x00000000, the last write
// time of the file MUST NOT be changed." Without that, a client changing only the
// read-only bit would drag the write time back to 1970.
func TestSetInformationAppliesAttributesAndWriteTime(t *testing.T) {
	fs, client := legacyInfoServer(t, false)

	stamp := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)

	request := commands.NewSetInformationRequest()
	pathRequest(t, &request.FileName, "described.txt")
	request.FileAttributes.SetAttributes(smbFileAttributeReadOnly)
	request.LastWriteTime = types.ULONG(utimeOf(stamp))

	status, _ := sendLegacy(t, client, codes.SMB_COM_SET_INFORMATION, request)
	if status != 0 {
		t.Fatalf("the set reported 0x%08X, want success", status)
	}

	attr, err := fs.Stat("described.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !attr.ReadOnly {
		t.Error("the read-only attribute was not applied")
	}
	if got := utimeOf(attr.Modified); got != utimeOf(stamp) {
		t.Errorf("the write time is %d, want %d", got, utimeOf(stamp))
	}

	// Now clear the read-only bit and send no write time; the stamp must hold.
	clearing := commands.NewSetInformationRequest()
	pathRequest(t, &clearing.FileName, "described.txt")
	clearing.FileAttributes.SetAttributes(0)
	clearing.LastWriteTime = types.ULONG(0)

	status, _ = sendLegacy(t, client, codes.SMB_COM_SET_INFORMATION, clearing)
	if status != 0 {
		t.Fatalf("the second set reported 0x%08X, want success", status)
	}

	attr, err = fs.Stat("described.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if attr.ReadOnly {
		t.Error("the read-only attribute was not cleared")
	}
	if got := utimeOf(attr.Modified); got != utimeOf(stamp) {
		t.Errorf("a zero write time moved the stamp to %d, want it left at %d", got, utimeOf(stamp))
	}
}

// TestSetInformationOnAReadOnlyShareIsRefused asserts the share's own refusal
// applies, as it does to every other modifying command.
func TestSetInformationOnAReadOnlyShareIsRefused(t *testing.T) {
	_, client := legacyInfoServer(t, true)

	request := commands.NewSetInformationRequest()
	pathRequest(t, &request.FileName, "described.txt")
	request.FileAttributes.SetAttributes(smbFileAttributeReadOnly)

	status, _ := sendLegacy(t, client, codes.SMB_COM_SET_INFORMATION, request)
	if status == 0 {
		t.Error("a set on a read-only share succeeded")
	}
}

// TestSetInformationOnAMissingFileIsRefused asserts a set does not create the file
// it names.
//
// [MS-CIFS] section 3.3.5.12 requires the file to exist. A backend asked to set
// attributes on a missing path might otherwise create it, which would make a
// metadata command a way to create files.
func TestSetInformationOnAMissingFileIsRefused(t *testing.T) {
	fs, client := legacyInfoServer(t, false)

	request := commands.NewSetInformationRequest()
	pathRequest(t, &request.FileName, "invented.txt")
	request.FileAttributes.SetAttributes(smbFileAttributeReadOnly)

	status, _ := sendLegacy(t, client, codes.SMB_COM_SET_INFORMATION, request)
	if status != uint32(nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND) {
		t.Errorf("setting on a missing file reported 0x%08X, want STATUS_OBJECT_NAME_NOT_FOUND", status)
	}
	if _, err := fs.Stat("invented.txt"); err == nil {
		t.Error("the refused set created the file")
	}
}

// TestQueryAndSetInformation2RoundTripTheTimestamps asserts the handle-based pair
// agree with each other.
//
// The timestamps travel as MS-DOS date and time pairs, which have two-second
// resolution, so the comparison is to the rounded value rather than to the one
// sent — a mismatch there would be the format, not the server.
func TestQueryAndSetInformation2RoundTripTheTimestamps(t *testing.T) {
	_, client := legacyInfoServer(t, false)

	fid, err := client.OpenFile("described.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}
	defer client.CloseFile(fid)

	// An odd second, so the two-second rounding is visible and accounted for.
	written := time.Date(2001, 2, 3, 4, 5, 7, 0, time.UTC)

	setRequest := commands.NewSetInformation2Request()
	setRequest.FID = types.USHORT(fid)
	setRequest.LastWriteDate, setRequest.LastWriteTime = dosDateTimeOf(written)

	status, _ := sendLegacy(t, client, codes.SMB_COM_SET_INFORMATION2, setRequest)
	if status != 0 {
		t.Fatalf("the set reported 0x%08X, want success", status)
	}

	queryRequest := commands.NewQueryInformation2Request()
	queryRequest.FID = types.USHORT(fid)

	status, body := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION2, queryRequest)
	if status != 0 {
		t.Fatalf("the query reported 0x%08X, want success", status)
	}

	response := commands.NewQueryInformation2Response()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}

	got := timeFromDosDateTime(response.LastWriteDate, response.LastWriteTime)
	want := timeFromDosDateTime(setRequest.LastWriteDate, setRequest.LastWriteTime)
	if !got.Equal(want) {
		t.Errorf("the write time came back as %s, want %s", got, want)
	}

	if int(response.FileDataSize) != len("0123456789") {
		t.Errorf("FileDataSize is %d, want %d", response.FileDataSize, len("0123456789"))
	}
}

// TestQueryInformation2OnAPipeHandleIsRefused asserts the FID has to name a
// regular file.
//
// [MS-CIFS] section 3.3.5.28: "The FID MUST indicate a regular file." A pipe
// handle has no timestamps or size to report, so answering would mean inventing
// them.
func TestQueryInformation2OnAPipeHandleIsRefused(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	request := commands.NewQueryInformation2Request()
	request.FID = types.USHORT(fid)

	status, _ := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION2, request)
	if status == 0 {
		t.Error("querying a pipe handle for file information succeeded")
	}

	if _, err := client.Echo([]byte("alive")); err != nil {
		t.Fatalf("the connection did not survive the refusal: %v", err)
	}
}

// TestLegacyInfoOnAnUnknownHandleIsRefused asserts a FID the connection never
// issued is reported rather than dereferenced.
func TestLegacyInfoOnAnUnknownHandleIsRefused(t *testing.T) {
	_, client := legacyInfoServer(t, false)

	request := commands.NewQueryInformation2Request()
	request.FID = types.USHORT(0xBEEF)

	status, _ := sendLegacy(t, client, codes.SMB_COM_QUERY_INFORMATION2, request)
	if status != uint32(nt_status.NT_STATUS_SMB_BAD_FID) {
		t.Errorf("an unknown FID reported 0x%08X, want STATUS_SMB_BAD_FID (0x%08X)",
			status, uint32(nt_status.NT_STATUS_SMB_BAD_FID))
	}

	if _, err := client.Echo([]byte("alive")); err != nil {
		t.Fatalf("the connection did not survive the refusal: %v", err)
	}
}
