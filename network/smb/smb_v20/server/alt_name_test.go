package server_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// TestAlternateNameInformationHasNoShortName checks that a QUERY_INFO for
// FileAlternateNameInformation is answered STATUS_OBJECT_NAME_NOT_FOUND rather
// than with the long name. The served volume generates no 8.3 short names, and
// Windows reports that absence with OBJECT_NAME_NOT_FOUND; returning the long
// name (which is never 8.3) was a server fingerprint (M-090 / #2059).
func TestAlternateNameInformationHasNoShortName(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	id, err := c.CreateFile("a long file name.txt", fileflags.GENERIC_READ,
		fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN_IF, 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// FileAlternateNameInformation must be refused: there is no short name.
	altStatus := queryInfoStatus(t, c, id, uint8(infoclass.FileAlternateNameInformation))
	if altStatus != uint32(nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND) {
		t.Errorf("FileAlternateNameInformation status 0x%08x, want STATUS_OBJECT_NAME_NOT_FOUND", altStatus)
	}

	// FileNameInformation still answers with the long name, unchanged.
	nameStatus := queryInfoStatus(t, c, id, uint8(infoclass.FileNameInformation))
	if nameStatus != uint32(nt_status.NT_STATUS_SUCCESS) {
		t.Errorf("FileNameInformation status 0x%08x, want STATUS_SUCCESS", nameStatus)
	}
}

// queryInfoStatus issues a single QUERY_INFO for a file information class and
// returns the raw status the server answered. An ECHO is appended because an
// unrelated compound needs at least two requests.
func queryInfoStatus(t *testing.T, c *client.Client, id types.SMB2_FILEID, class uint8) uint32 {
	t.Helper()
	req := commands.NewQueryInfoRequest()
	req.InfoType = commands.SMB2_0_INFO_FILE
	req.FileInfoClass = types.UCHAR(class)
	req.FileId = id
	req.OutputBufferLength = 4096
	answers, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(req), c.NewRequest(commands.NewEchoRequest())})
	if err != nil {
		t.Fatalf("query info class %d: %v", class, err)
	}
	return answers[0].Header.Status
}
