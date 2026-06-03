package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestTransaction2RequestSetupMarshal verifies that the Setup word array is
// marshalled with SetupCount derived from its length and each word encoded as a
// little-endian USHORT, per [MS-CIFS] 2.2.4.46.1. The first setup word carries
// the TRANS2 subcommand (e.g. TRANS2_FIND_FIRST2).
func TestTransaction2RequestSetupMarshal(t *testing.T) {
	out := commands.NewTransaction2Request()
	out.Setup = []types.USHORT{0x0001, 0x1234} // TRANS2_FIND_FIRST2 + a second word

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// SetupCount must be derived from len(Setup).
	if out.SetupCount != types.UCHAR(len(out.Setup)) {
		t.Errorf("SetupCount = %d, want %d", out.SetupCount, len(out.Setup))
	}

	// The two setup words must appear in the parameter block as little-endian USHORTs.
	wantWords := []byte{0x01, 0x00, 0x34, 0x12}
	if !bytes.Contains(marshalled, wantWords) {
		t.Errorf("marshalled output missing little-endian Setup words % x\nfull = % x", wantWords, marshalled)
	}

	// A request with no setup words must not contain them.
	empty := commands.NewTransaction2Request()
	emptyMarshalled, err := empty.Marshal()
	if err != nil {
		t.Fatalf("Marshal (empty) failed: %v", err)
	}
	if len(marshalled) != len(emptyMarshalled)+4 {
		t.Errorf("expected 4 extra bytes for 2 setup words, got delta %d", len(marshalled)-len(emptyMarshalled))
	}
}
