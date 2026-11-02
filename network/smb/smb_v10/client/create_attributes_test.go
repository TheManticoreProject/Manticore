package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// TestOpenFileWithAttributesSendsThem checks the caller's FILE_ATTRIBUTE_* reach
// the NT_CREATE_ANDX ExtFileAttributes, and that zero still sends
// FILE_ATTRIBUTE_NORMAL.
func TestOpenFileWithAttributesSendsThem(t *testing.T) {
	for _, tc := range []struct{ attrs, want uint32 }{
		{0x00000002 | 0x00000001, 0x00000003}, // HIDDEN | READONLY
		{0, 0x00000080},                       // NORMAL
	} {
		resp := commands.NewNtCreateAndxResponse()
		resp.FID = 0x4242
		tr := &capturingTransport{response: marshalResponse(t, resp)}
		c := newSessionClient(tr)

		if _, err := c.OpenFileWithAttributes("f.txt", 0x0012019f, 0x07, 0x00000005, 0x00000040, tc.attrs); err != nil {
			t.Fatalf("OpenFileWithAttributes: %v", err)
		}
		reqMsg := message.NewMessage()
		if err := reqMsg.Unmarshal(tr.sent); err != nil {
			t.Fatalf("unmarshal captured request: %v", err)
		}
		req, ok := reqMsg.Command.(*commands.NtCreateAndxRequest)
		if !ok {
			t.Fatalf("captured command is %T, want *NtCreateAndxRequest", reqMsg.Command)
		}
		if got := uint32(req.ExtFileAttributes); got != tc.want {
			t.Errorf("attrs 0x%x: ExtFileAttributes = 0x%08x, want 0x%08x", tc.attrs, got, tc.want)
		}
	}
}
