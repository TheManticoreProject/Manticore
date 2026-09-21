package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/createcontext"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func createResponseWithDH2Q(t *testing.T, fileId types.SMB2_FILEID, timeout, flags uint32) []byte {
	t.Helper()
	resp := commands.NewCreateResponse()
	resp.FileId = fileId
	resp.OplockLevel = commands.SMB2_OPLOCK_LEVEL_BATCH

	dh2qResp := make([]byte, 8)
	binary.LittleEndian.PutUint32(dh2qResp[0:4], timeout)
	binary.LittleEndian.PutUint32(dh2qResp[4:8], flags)
	ctx := createcontext.CreateContext{
		Name: createcontext.NameDurableHandleReqV2,
		Data: dh2qResp,
	}
	ctxBuf, err := createcontext.Marshal([]createcontext.CreateContext{ctx})
	if err != nil {
		t.Fatalf("marshal DH2Q response context: %v", err)
	}
	resp.CreateContexts = ctxBuf

	return cannedResponse(t, resp, 0, 0x5, 0x99)
}

func TestCreateFileWithDurableHandleV2(t *testing.T) {
	fileId := types.SMB2_FILEID{Persistent: 0xAAAA, Volatile: 0xBBBB}
	ft := &fakeTransport{
		connected: true,
		responses: [][]byte{createResponseWithDH2Q(t, fileId, 30000, 0)},
	}
	c := withConnectedTree(ft)

	dh, err := c.CreateFileWithDurableHandleV2("test.txt", 0x001F01FF, 0x07, 0x01, 0x00000040, 30000, false)
	if err != nil {
		t.Fatalf("CreateFileWithDurableHandleV2: %v", err)
	}
	if dh.FileId.Persistent != 0xAAAA || dh.FileId.Volatile != 0xBBBB {
		t.Errorf("FileId = %+v, want {AAAA BBBB}", dh.FileId)
	}
	if dh.Timeout != 30000 {
		t.Errorf("Timeout = %d, want 30000", dh.Timeout)
	}
	if dh.IsPersistent() {
		t.Error("IsPersistent should be false")
	}
	zero := [16]byte{}
	if dh.CreateGuid == zero {
		t.Error("CreateGuid should not be all zeros")
	}
}

func TestCreateFileWithDurableHandleV2Persistent(t *testing.T) {
	fileId := types.SMB2_FILEID{Persistent: 0x1111, Volatile: 0x2222}
	ft := &fakeTransport{
		connected: true,
		responses: [][]byte{createResponseWithDH2Q(t, fileId, 60000, createcontext.SMB2_DHANDLE_FLAG_PERSISTENT)},
	}
	c := withConnectedTree(ft)

	dh, err := c.CreateFileWithDurableHandleV2("share\\file.dat", 0x001F01FF, 0x07, 0x01, 0, 60000, true)
	if err != nil {
		t.Fatalf("CreateFileWithDurableHandleV2: %v", err)
	}
	if !dh.IsPersistent() {
		t.Error("IsPersistent should be true when server echoes persistent flag")
	}
	if dh.Timeout != 60000 {
		t.Errorf("Timeout = %d, want 60000", dh.Timeout)
	}
}

func TestReconnectDurableHandleV2(t *testing.T) {
	newFileId := types.SMB2_FILEID{Persistent: 0xAAAA, Volatile: 0xDDDD}
	reconResp := commands.NewCreateResponse()
	reconResp.FileId = newFileId

	ft := &fakeTransport{
		connected: true,
		responses: [][]byte{cannedResponse(t, reconResp, 0, 0x5, 0x99)},
	}
	c := withConnectedTree(ft)

	dh := &DurableHandle{
		FileId:     types.SMB2_FILEID{Persistent: 0xAAAA, Volatile: 0xBBBB},
		CreateGuid: [16]byte{0x01, 0x02, 0x03, 0x04},
	}

	fid, err := c.ReconnectDurableHandleV2(dh)
	if err != nil {
		t.Fatalf("ReconnectDurableHandleV2: %v", err)
	}
	if fid.Persistent != 0xAAAA {
		t.Errorf("Persistent = 0x%x, want 0xAAAA", fid.Persistent)
	}
	if fid.Volatile != 0xDDDD {
		t.Errorf("Volatile = 0x%x, want 0xDDDD", fid.Volatile)
	}

	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
	}
	sent := ft.sent[0]
	// The create context data starts after the fixed header + create body.
	// Verify it carries the DH2C name tag somewhere in the wire bytes.
	found := false
	for i := 0; i+4 <= len(sent); i++ {
		if string(sent[i:i+4]) == "DH2C" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DH2C name tag not found in sent frame")
	}
}

func TestCreateFileWithDurableHandleV2RequiresTree(t *testing.T) {
	ft := &fakeTransport{connected: true}
	c := newTestClient(ft)

	_, err := c.CreateFileWithDurableHandleV2("file.txt", 0, 0, 0, 0, 0, false)
	if err == nil {
		t.Error("expected error when no tree connect is established")
	}
}

func TestReconnectDurableHandleV2RequiresTree(t *testing.T) {
	ft := &fakeTransport{connected: true}
	c := newTestClient(ft)

	_, err := c.ReconnectDurableHandleV2(&DurableHandle{})
	if err == nil {
		t.Error("expected error when no tree connect is established")
	}
}
