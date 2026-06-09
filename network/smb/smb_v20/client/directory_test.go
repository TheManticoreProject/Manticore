package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestQueryDirectory(t *testing.T) {
	qdResp := commands.NewQueryDirectoryResponse()
	qdResp.OutputBuffer = []byte{0x60, 0x00, 0x00, 0x00, 0x11, 0x22}

	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, qdResp, 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	entries, err := c.QueryDirectory(types.SMB2_FILEID{Persistent: 0x1}, 0x25, "*", 0)
	if err != nil {
		t.Fatalf("QueryDirectory: %v", err)
	}
	if !bytes.Equal(entries, qdResp.OutputBuffer) {
		t.Errorf("entries = % x, want % x", entries, qdResp.OutputBuffer)
	}
}

func TestQueryDirectoryNoMoreFiles(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, commands.NewQueryDirectoryResponse(), ntStatusNoMoreFiles, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	entries, err := c.QueryDirectory(types.SMB2_FILEID{}, 0x25, "*", 0)
	if err != nil {
		t.Fatalf("QueryDirectory at end returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries at end = % x, want empty", entries)
	}
}
