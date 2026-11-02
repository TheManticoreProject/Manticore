package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// sentCreateAttributes decodes the FileAttributes of the CREATE request the
// client sent.
func sentCreateAttributes(t *testing.T, ft *fakeTransport) uint32 {
	t.Helper()
	if len(ft.sent) == 0 {
		t.Fatal("no request sent")
	}
	req := commands.NewCreateRequest()
	if _, err := req.Unmarshal(ft.sent[0][64:]); err != nil {
		t.Fatalf("decode CREATE request: %v", err)
	}
	return req.FileAttributes
}

// TestCreateFileWithAttributesSendsThem checks the caller's FILE_ATTRIBUTE_*
// reach the CREATE request, and that zero still sends FILE_ATTRIBUTE_NORMAL.
func TestCreateFileWithAttributesSendsThem(t *testing.T) {
	for _, tc := range []struct{ attrs, want uint32 }{
		{0x00000002 | 0x00000004, 0x00000006}, // HIDDEN | SYSTEM
		{0, 0x00000080},                       // NORMAL
	} {
		createResp := commands.NewCreateResponse()
		createResp.FileId = types.SMB2_FILEID{Persistent: 1, Volatile: 2}
		ft := &fakeTransport{responses: [][]byte{cannedResponse(t, createResp, 0, 0x5, 0x99)}}
		c := withConnectedTree(ft)
		if _, err := c.CreateFileWithAttributes("f.txt", 0x00100081, 0x07, 0x00000005, 0x00000040, tc.attrs); err != nil {
			t.Fatalf("CreateFileWithAttributes: %v", err)
		}
		if got := sentCreateAttributes(t, ft); got != tc.want {
			t.Errorf("attrs 0x%x: CREATE FileAttributes = 0x%08x, want 0x%08x", tc.attrs, got, tc.want)
		}
	}
}
