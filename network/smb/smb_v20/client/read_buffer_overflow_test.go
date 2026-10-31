package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TestReadBufferOverflowReturnsPartialMessage verifies that a READ answered with
// STATUS_BUFFER_OVERFLOW — a named-pipe message larger than the read — returns
// the data that did fit instead of failing, and that ReadFile stops there rather
// than reading on into the remainder: one READ is sent even though fewer than
// the requested bytes were returned.
func TestReadBufferOverflowReturnsPartialMessage(t *testing.T) {
	partial := commands.NewReadResponse()
	partial.Data = []byte("part")
	rest := commands.NewReadResponse()
	rest.Data = []byte("rest")
	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, partial, ntStatusBufferOverflow, 0x5, 0x99),
		cannedResponse(t, rest, 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)
	// Split the 8-byte read into two 4-byte READs.
	c.Connection.Server.MaxReadSize = 4

	data, err := c.ReadFile(types.SMB2_FILEID{}, 0, 8)
	if err != nil {
		t.Fatalf("ReadFile with STATUS_BUFFER_OVERFLOW returned error: %v", err)
	}
	if string(data) != "part" {
		t.Errorf("ReadFile = %q, want %q", data, "part")
	}
	if len(ft.sent) != 1 {
		t.Errorf("ReadFile sent %d READ requests, want 1", len(ft.sent))
	}
}
