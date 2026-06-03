package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// TestNegotiateResponseUnmarshalNonExtendedNoPanic reproduces the slice-bounds
// panic that occurred when unmarshalling a non-extended-security NegotiateResponse
// whose data block is empty: GetNullTerminatedUnicodeString returns an offset of
// len(string)+2, which exceeded the remaining bytes and caused a [2:0] slice on
// the DomainName/ServerName reads. The Unmarshal must clamp the offset and not panic.
func TestNegotiateResponseUnmarshalNonExtendedNoPanic(t *testing.T) {
	// Default response: Capabilities has no CAP_EXTENDED_SECURITY bit, so the
	// non-extended (Challenge + DomainName + ServerName) data layout is used.
	out := commands.NewNegotiateResponse()

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unmarshal panicked: %v", r)
		}
	}()

	in := commands.NewNegotiateResponse()
	if _, err := in.Unmarshal(marshalled); err != nil {
		t.Logf("Unmarshal returned error (acceptable): %v", err)
	}
}
