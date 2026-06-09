package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// withConnectedTree returns a client with a session and selected tree installed,
// so file operations can be exercised over the in-memory transport.
func withConnectedTree(ft *fakeTransport) *Client {
	c := newTestClient(ft)
	c.Session = &Session{Client: c, SessionId: 0x99, TreeId: 0x5}
	c.Connection.SessionTable[0x99] = c.Session
	return c
}

func TestCreateReadWriteClose(t *testing.T) {
	createResp := commands.NewCreateResponse()
	createResp.FileId = types.SMB2_FILEID{Persistent: 0xCAFE, Volatile: 0xBEEF}

	writeResp := commands.NewWriteResponse()
	writeResp.Count = 5

	readResp := commands.NewReadResponse()
	readResp.Data = []byte("hello")

	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, createResp, 0, 0x5, 0x99),
		cannedResponse(t, writeResp, 0, 0x5, 0x99),
		cannedResponse(t, readResp, 0, 0x5, 0x99),
		cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	fileId, err := c.CreateFile("dir\\file.txt", 0x00100081, 0x07, 0x00000005, 0x00000040)
	if err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	if fileId.Persistent != 0xCAFE || fileId.Volatile != 0xBEEF {
		t.Errorf("FileId = %+v, want {CAFE BEEF}", fileId)
	}

	n, err := c.WriteFile(fileId, 0, []byte("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteFile count = %d, want 5", n)
	}

	data, err := c.ReadFile(fileId, 0, 1024)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(data, []byte("hello")) {
		t.Errorf("ReadFile = %q, want %q", data, "hello")
	}

	if err := c.CloseFile(fileId); err != nil {
		t.Fatalf("CloseFile: %v", err)
	}
}

func TestReadEndOfFileReturnsEmpty(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, commands.NewReadResponse(), ntStatusEndOfFile, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	data, err := c.ReadFile(types.SMB2_FILEID{}, 0x1000, 512)
	if err != nil {
		t.Fatalf("ReadFile at EOF returned error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("ReadFile at EOF = %q, want empty", data)
	}
}

func TestCreateFileRequiresTree(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Session = &Session{Client: c, SessionId: 0x99} // session but no tree (TreeId 0)
	if _, err := c.CreateFile("x", 0, 0, 0, 0); err == nil {
		t.Errorf("expected CreateFile without a tree connect to error")
	}
}
