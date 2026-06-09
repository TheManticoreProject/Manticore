package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

func TestSignAndVerifyRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef") // 16-byte session key

	m := message.NewMessage()
	m.Header.MessageId = 3
	m.SetCommand(commands.NewTreeDisconnectRequest())
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	signMessage(key, wire)

	// The SIGNED flag must be set after signing.
	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Header.Flags.IsSigned() {
		t.Errorf("SMB2_FLAGS_SIGNED not set after signing")
	}

	// The signature must verify with the same key and fail with a different key.
	if !verifySignature(key, wire) {
		t.Errorf("verifySignature failed for a correctly signed message")
	}
	if verifySignature([]byte("WRONGKEYWRONGKEY"), wire) {
		t.Errorf("verifySignature succeeded with the wrong key")
	}

	// Tampering with the body must invalidate the signature.
	wire[len(wire)-1] ^= 0xFF
	if verifySignature(key, wire) {
		t.Errorf("verifySignature succeeded for a tampered message")
	}
}

func TestVerifyRejectsShortMessage(t *testing.T) {
	if verifySignature([]byte("key"), make([]byte, 10)) {
		t.Errorf("verifySignature should reject a sub-header-length message")
	}
}
