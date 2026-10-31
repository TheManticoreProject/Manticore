package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/fingerprint"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestNegotiateOffersNoCompressionByDefault verifies that the default profile's
// SMB 3.1.1 NEGOTIATE carries no SMB2_COMPRESSION_CAPABILITIES context. The
// client cannot decode an SMB2 COMPRESSION_TRANSFORM_HEADER (0xFC 'S' 'M' 'B')
// response, so offering algorithms would let a server compress replies the
// client then rejects.
func TestNegotiateOffersNoCompressionByDefault(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponse(t)}}
	c := newTestClient(ft)
	c.Profile = fingerprint.Default()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_3_1_1}

	if err := c.Negotiate(); err != nil {
		t.Fatalf("Negotiate: %v", err)
	}
	sent := message.NewMessage()
	if _, err := sent.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("parsing sent NEGOTIATE: %v", err)
	}
	req, ok := sent.Command.(*commands.NegotiateRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *NegotiateRequest", sent.Command)
	}
	if len(req.Contexts) == 0 {
		t.Fatal("SMB 3.1.1 NEGOTIATE carried no negotiate contexts")
	}
	for _, ctx := range req.Contexts {
		if ctx.ContextType == commands.SMB2_COMPRESSION_CAPABILITIES {
			t.Errorf("NEGOTIATE offers compression (context data % x), which the client cannot decode", ctx.Data)
		}
	}
}
