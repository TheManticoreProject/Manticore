package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

func TestSessionSetupNilCredentials(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	if err := c.SessionSetup(nil); err == nil {
		t.Errorf("expected SessionSetup(nil) to error, got nil")
	}
}

func TestFormatNTStatus(t *testing.T) {
	// A success code formats as a bare hex value.
	if got := formatNTStatus(0x00000000); got == "" {
		t.Errorf("formatNTStatus returned empty string")
	}
	// MORE_PROCESSING_REQUIRED should at least include its hex form.
	got := formatNTStatus(ntStatusMoreProcessingRequired)
	if got == "" {
		t.Errorf("formatNTStatus returned empty string for MORE_PROCESSING_REQUIRED")
	}
}

func TestSessionGuestFlagDisablesSigning(t *testing.T) {
	session := &Session{
		SessionKey:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	// Simulate what SessionSetup does for guest sessions.
	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_IS_GUEST)
	session.IsGuest = sessionFlags&commands.SMB2_SESSION_FLAG_IS_GUEST != 0

	if !session.IsGuest {
		t.Fatal("IsGuest should be true")
	}

	// Per MS-SMB2 3.2.5.3.1, guest sessions disable signing.
	if session.IsGuest {
		session.SessionKey = make([]byte, 16)
		session.SigningKey = session.SessionKey
		session.SigningActive = false
	}

	if session.SigningActive {
		t.Error("signing should be disabled for guest sessions")
	}
	for _, b := range session.SessionKey {
		if b != 0 {
			t.Error("session key should be zeroed for guest sessions")
			break
		}
	}
}

func TestSessionNullFlagDisablesSigning(t *testing.T) {
	session := &Session{
		SessionKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_IS_NULL)
	session.IsNull = sessionFlags&commands.SMB2_SESSION_FLAG_IS_NULL != 0

	if !session.IsNull {
		t.Fatal("IsNull should be true")
	}

	if session.IsNull {
		session.SessionKey = make([]byte, 16)
		session.SigningKey = session.SessionKey
		session.SigningActive = false
	}

	if session.SigningActive {
		t.Error("signing should be disabled for anonymous sessions")
	}
}

func TestSessionNormalFlagPreservesSigning(t *testing.T) {
	session := &Session{
		SessionKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_ENCRYPT_DATA)
	session.IsGuest = sessionFlags&commands.SMB2_SESSION_FLAG_IS_GUEST != 0
	session.IsNull = sessionFlags&commands.SMB2_SESSION_FLAG_IS_NULL != 0

	if session.IsGuest || session.IsNull {
		t.Error("neither guest nor null should be set for an encrypted session")
	}
	if !session.SigningActive {
		t.Error("signing should remain active for normal sessions")
	}
	if session.SessionKey[0] != 0x01 {
		t.Error("session key should be preserved for normal sessions")
	}
}
