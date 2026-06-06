package message_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// roundTripSessionSetupResponse marshals a SessionSetupAndxResponse carrying the
// given NativeOS/NativeLanMan bytes inside a reply message whose Flags2 has (or
// lacks) the Unicode bit, then unmarshals it back and returns the decoded command.
func roundTripSessionSetupResponse(t *testing.T, unicode bool, nativeOS, nativeLanMan []byte) *commands.SessionSetupAndxResponse {
	t.Helper()

	msg := message.NewMessage()
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	if unicode {
		msg.Header.Flags2 = flags2.FLAGS2_UNICODE
	}

	resp := commands.NewSessionSetupAndxResponse()
	resp.NativeOS = []types.UCHAR(nativeOS)
	resp.NativeLanMan = []types.UCHAR(nativeLanMan)
	msg.AddCommand(resp)

	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	out := message.NewMessage()
	if err := out.Unmarshal(raw); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	decoded, ok := out.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		t.Fatalf("unexpected command type %T", out.Command)
	}
	return decoded
}

// TestSessionSetupResponseDecodesOEMStrings verifies that, when the message does
// not set SMB_FLAGS2_UNICODE, the NativeOS/NativeLanMan strings are decoded as
// OEM (single-byte) strings rather than UTF-16.
func TestSessionSetupResponseDecodesOEMStrings(t *testing.T) {
	decoded := roundTripSessionSetupResponse(t, false, []byte("Unix\x00"), []byte("Samba\x00"))

	if got := string(decoded.NativeOS); got != "Unix" {
		t.Errorf("expected NativeOS %q, got %q", "Unix", got)
	}
	if got := string(decoded.NativeLanMan); got != "Samba" {
		t.Errorf("expected NativeLanMan %q, got %q", "Samba", got)
	}
}

// TestSessionSetupResponseDecodesUnicodeStrings verifies that, when the message
// sets SMB_FLAGS2_UNICODE, the NativeOS/NativeLanMan strings are decoded as
// UTF-16LE byte runs (read up to the double-null terminator).
func TestSessionSetupResponseDecodesUnicodeStrings(t *testing.T) {
	// UTF-16LE "Win" and "SMB" followed by a 2-byte null terminator.
	nativeOS := []byte{'W', 0, 'i', 0, 'n', 0, 0, 0}
	nativeLanMan := []byte{'S', 0, 'M', 0, 'B', 0, 0, 0}

	decoded := roundTripSessionSetupResponse(t, true, nativeOS, nativeLanMan)

	if want := []byte{'W', 0, 'i', 0, 'n', 0}; !bytes.Equal(decoded.NativeOS, want) {
		t.Errorf("expected NativeOS bytes %v, got %v", want, []byte(decoded.NativeOS))
	}
	if want := []byte{'S', 0, 'M', 0, 'B', 0}; !bytes.Equal(decoded.NativeLanMan, want) {
		t.Errorf("expected NativeLanMan bytes %v, got %v", want, []byte(decoded.NativeLanMan))
	}
}
