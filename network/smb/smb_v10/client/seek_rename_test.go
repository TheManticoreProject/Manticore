package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
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

// On a Unicode connection the header carries FLAGS2_UNICODE, so NT_RENAME's old and
// new names must be UTF-16LE; UTF-8 bytes would be decoded by the server as unrelated
// UTF-16 characters.
func TestNtRenameUsesUnicodeNames(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(c *client.Client) error
	}{
		{"NtRename", func(c *client.Client) error { return c.NtRename(`\old.txt`, `\new.txt`) }},
		{"CreateHardLink", func(c *client.Client) error { return c.CreateHardLink(`\old.txt`, `\new.txt`) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := &capturingTransport{response: marshalResponse(t, commands.NewNtRenameResponse())}
			c := newSessionClient(tr)
			c.Connection.Server.Capabilities = capabilities.CAP_UNICODE
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}

			m := message.NewMessage()
			if err := m.Unmarshal(tr.sent); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			req := m.Command.(*commands.NtRenameRequest)
			if got := req.OldFileName.StringWithEncoding(true); got != `\old.txt` {
				t.Errorf("OldFileName decodes as %q, want %q", got, `\old.txt`)
			}
			if got := req.NewFileName.StringWithEncoding(true); got != `\new.txt` {
				t.Errorf("NewFileName decodes as %q, want %q", got, `\new.txt`)
			}
		})
	}
}
