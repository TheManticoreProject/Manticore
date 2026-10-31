package negotiate_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// TestNegotiateOEMFieldsUnderUnicode checks that DomainName and WorkstationName are
// written in the OEM character set even when NTLMSSP_NEGOTIATE_UNICODE is set: the
// NEGOTIATE precedes the character-set negotiation, and [MS-NLMP] 2.2.1.1 requires
// both payload fields to be OEM-encoded.
func TestNegotiateOEMFieldsUnderUnicode(t *testing.T) {
	negotiateFlags := flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_NTLM | flags.NTLMSSP_REQUEST_TARGET

	msg, err := negotiate.CreateNegotiateMessage("corp", "ws01", negotiateFlags, nil)
	if err != nil {
		t.Fatalf("CreateNegotiateMessage: %v", err)
	}
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	field := func(offset int) []byte {
		length := int(binary.LittleEndian.Uint16(raw[offset : offset+2]))
		start := int(binary.LittleEndian.Uint32(raw[offset+4 : offset+8]))
		if start+length > len(raw) {
			t.Fatalf("field at %d runs past the message: offset %d, length %d, message %d", offset, start, length, len(raw))
		}
		return raw[start : start+length]
	}
	if got := field(16); !bytes.Equal(got, []byte("CORP")) {
		t.Errorf("DomainName on the wire = % x, want OEM % x", got, []byte("CORP"))
	}
	if got := field(24); !bytes.Equal(got, []byte("WS01")) {
		t.Errorf("WorkstationName on the wire = % x, want OEM % x", got, []byte("WS01"))
	}

	var parsed negotiate.NegotiateMessage
	if _, err := parsed.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := parsed.GetDomainName(); got != "CORP" {
		t.Errorf("GetDomainName() = %q, want %q", got, "CORP")
	}
	if got := parsed.GetWorkstationName(); got != "WS01" {
		t.Errorf("GetWorkstationName() = %q, want %q", got, "WS01")
	}

	parsed.SetWorkstationName("other")
	if !bytes.Equal(parsed.Workstation, []byte("OTHER")) {
		t.Errorf("SetWorkstationName under UNICODE stored % x, want OEM % x", parsed.Workstation, []byte("OTHER"))
	}
}
