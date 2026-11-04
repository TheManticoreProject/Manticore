package client

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

// cannedChallengeResponse builds the wire bytes of a SESSION_SETUP response
// carrying STATUS_MORE_PROCESSING_REQUIRED and a stand-in CHALLENGE buffer.
func cannedChallengeResponse(t *testing.T) []byte {
	t.Helper()
	resp := commands.NewSessionSetupResponse()
	resp.SecurityBuffer = []byte{0x01, 0x02, 0x03, 0x04}

	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.Status = ntStatusMoreProcessingRequired
	m.SetCommand(resp)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building canned challenge response: %v", err)
	}
	return wire
}

// TestRelayNegotiateLegIsUnsignedDespiteExistingSession asserts that the opening
// leg of an outbound relay goes out on SessionId 0, unsigned, even when the
// client already holds an established, signing-active session. The engine keys
// signing off c.Session, so without clearing it the SessionId-0 request would be
// signed with the unrelated session's key — a frame a server rejects.
func TestRelayNegotiateLegIsUnsignedDespiteExistingSession(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedChallengeResponse(t)}}
	c := newTestClient(ft)
	c.Connection.Dialect = dialects.SMB2_DIALECT_2_1_0

	// A pre-existing, signing-active session on this client.
	c.Session = &Session{
		Client:        c,
		SessionId:     0x1122334455667788,
		SessionKey:    bytes.Repeat([]byte{0xAB}, 16),
		SigningKey:    bytes.Repeat([]byte{0xAB}, 16),
		SigningActive: true,
	}

	if _, _, err := c.SessionSetupNegotiateBlob([]byte{0x4e, 0x54, 0x4c, 0x4d}); err != nil {
		t.Fatalf("SessionSetupNegotiateBlob: %v", err)
	}

	sent := ft.sent[len(ft.sent)-1]
	if len(sent) < 64 {
		t.Fatalf("sent frame is %d bytes, too short for an SMB2 header", len(sent))
	}
	if sessionId := binary.LittleEndian.Uint64(sent[40:48]); sessionId != 0 {
		t.Errorf("opening leg SessionId = %#x, want 0", sessionId)
	}
	if signed := binary.LittleEndian.Uint32(sent[16:20]) & uint32(flags.SMB2_FLAGS_SIGNED); signed != 0 {
		t.Errorf("opening leg carries SMB2_FLAGS_SIGNED, want it unset")
	}
	if sig := sent[48:64]; !bytes.Equal(sig, make([]byte, 16)) {
		t.Errorf("opening leg is signed with the stale session key: signature = % x, want all zero", sig)
	}
}
