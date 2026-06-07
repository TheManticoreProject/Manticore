package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSeekReturnsNewOffset(t *testing.T) {
	resp := commands.NewSeekResponse()
	resp.Offset = types.ULONG(0x12345)

	tr := &capturingTransport{response: marshalResponse(t, resp)}
	c := newSessionClient(tr)

	got, err := c.Seek(3, client.SeekModeEnd, -16)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	}
	if got != 0x12345 {
		t.Errorf("Seek returned 0x%x, want 0x12345", got)
	}

	// Verify the request carried the FID, mode, and signed offset.
	msg := message.NewMessage()
	if err := msg.Unmarshal(tr.sent); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	req, ok := msg.Command.(*commands.SeekRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *SeekRequest", msg.Command)
	}
	if req.FID != 3 || uint16(req.Mode) != client.SeekModeEnd || int32(req.Offset) != -16 {
		t.Errorf("unexpected request: FID=%d Mode=%d Offset=%d", req.FID, req.Mode, int32(req.Offset))
	}
}

func TestNtRenameRequestShape(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewNtRenameResponse())}
	c := newSessionClient(tr)

	if err := c.NtRename("\\a.txt", "\\b.txt"); err != nil {
		t.Fatalf("NtRename: %v", err)
	}
	req := sentNtRename(t, tr.sent)
	if uint16(req.InformationLevel) != 0x0104 {
		t.Errorf("expected InformationLevel 0x0104 (RENAME_FILE), got 0x%04x", req.InformationLevel)
	}
}

func TestCreateHardLinkRequestShape(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewNtRenameResponse())}
	c := newSessionClient(tr)

	if err := c.CreateHardLink("\\orig.txt", "\\link.txt"); err != nil {
		t.Fatalf("CreateHardLink: %v", err)
	}
	req := sentNtRename(t, tr.sent)
	if uint16(req.InformationLevel) != 0x0103 {
		t.Errorf("expected InformationLevel 0x0103 (SET_LINK_INFO), got 0x%04x", req.InformationLevel)
	}
}

func TestSeekAndRenameWithoutSession(t *testing.T) {
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if _, err := c.Seek(1, client.SeekModeStart, 0); err == nil {
		t.Error("expected error without a session (Seek)")
	}
	if err := c.NtRename("\\a", "\\b"); err == nil {
		t.Error("expected error without a session (NtRename)")
	}
}

func sentNtRename(t *testing.T, raw []byte) *commands.NtRenameRequest {
	t.Helper()
	msg := message.NewMessage()
	if err := msg.Unmarshal(raw); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	req, ok := msg.Command.(*commands.NtRenameRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *NtRenameRequest", msg.Command)
	}
	return req
}
