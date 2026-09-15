package commands

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
)

// TestSessionSetupUnicodeStringsAreAligned guards the defect: a Unicode NativeOS
// MUST start on a 2-byte boundary from the start of the SMB header
// ([MS-SMB] 2.2.4.6.1). The data block begins at an odd offset, so without an
// alignment byte a security blob of even length leaves the strings misaligned and
// a peer reads every character a byte out of phase — the identity the client
// reports comes out unintelligible.
//
// Both parities of the blob length are covered, because only one of them needs
// the pad and a fix that always padded would be just as wrong.
func TestSessionSetupUnicodeStringsAreAligned(t *testing.T) {
	const (
		nativeOS     = "Windows Server 2012 R2 Standard 9600"
		nativeLanMan = "Windows Server 2012 R2 Standard 6.3"
	)

	for _, blobLen := range []int{112, 113} {
		req := NewSessionSetupAndxRequest()
		req.SetUnicode(true)
		req.Capabilities = capabilities.CAP_UNICODE | capabilities.CAP_EXTENDED_SECURITY
		req.SecurityBlob = make([]byte, blobLen)
		req.NativeOS = nativeOS
		req.NativeLanMan = nativeLanMan

		raw, err := req.Marshal()
		if err != nil {
			t.Fatalf("blob %d: Marshal: %v", blobLen, err)
		}

		parsed := NewSessionSetupAndxRequest()
		parsed.SetUnicode(true)
		parsed.Capabilities = req.Capabilities
		if _, err := parsed.Unmarshal(raw); err != nil {
			t.Fatalf("blob %d: Unmarshal: %v", blobLen, err)
		}

		if parsed.NativeOS != nativeOS {
			t.Errorf("blob %d: NativeOS round-tripped as %q, want %q", blobLen, parsed.NativeOS, nativeOS)
		}
		if parsed.NativeLanMan != nativeLanMan {
			t.Errorf("blob %d: NativeLanMan round-tripped as %q, want %q", blobLen, parsed.NativeLanMan, nativeLanMan)
		}
	}
}

// TestSessionSetupOEMStringsAreNotPadded pins the other branch: an OEM message
// needs no alignment, and a pad byte inserted there would shift the strings by
// one in the opposite direction.
func TestSessionSetupOEMStringsAreNotPadded(t *testing.T) {
	const nativeOS = "PlainOS"

	req := NewSessionSetupAndxRequest()
	req.SetUnicode(false)
	req.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	req.SecurityBlob = make([]byte, 112)
	req.NativeOS = nativeOS
	req.NativeLanMan = "PlainLanMan"

	raw, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed := NewSessionSetupAndxRequest()
	parsed.SetUnicode(false)
	parsed.Capabilities = req.Capabilities
	if _, err := parsed.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if parsed.NativeOS != nativeOS {
		t.Errorf("NativeOS round-tripped as %q, want %q", parsed.NativeOS, nativeOS)
	}
}
