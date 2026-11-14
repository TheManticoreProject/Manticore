package dataqueue

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
)

// TestObjectEncodesNamesWithSystemCCSID pins that the data-queue client encodes a
// queue name and library through the session's sign-on system CCSID codec. The
// data-queue host server interprets names in the sign-on system CCSID, not the
// command-server job CCSID that rmtcmd uses, which matches the reference client
// (JTOpen BaseDataQueueImplRemote, getConverter(system_.getCcsid(), ...)). The
// three EBCDIC variant characters $, # and @ sit at different code points per
// CCSID, so a system-CCSID codec produces that CCSID's own bytes; if the path were
// switched to a different CCSID, or the variant table moved, these bytes would
// change and the server would see a different name. Settles #1965.
func TestObjectEncodesNamesWithSystemCCSID(t *testing.T) {
	// For each national CCSID, the expected trailing code points for "$#@" (the
	// rest of the 10-byte field is EBCDIC blanks, 0x40). Values match the shared
	// variant-character table (see ibmi/ccsid TestObjectNameVariantCharacters...).
	cases := []struct {
		id      ccsid.CCSID
		variant []byte // bytes for "$#@"
	}{
		{37, []byte{0x5B, 0x7B, 0x7C}},
		{297, []byte{0x5B, 0xB1, 0x44}}, // $ # @ differ from CCSID 37
		{273, []byte{0x5B, 0x7B, 0xB5}},
		{285, []byte{0x4A, 0x7B, 0x7C}},
	}
	for _, tc := range cases {
		codec, err := ccsid.Lookup(tc.id)
		if err != nil {
			t.Fatalf("Lookup(%d): %v", tc.id, err)
		}
		s := &Session{codec: codec}
		// A valid object name whose last three characters are the variant set.
		obj, err := s.object(Queue{Library: "QGPL", Name: "Q$#@"})
		if err != nil {
			t.Fatalf("object() CCSID %d: %v", tc.id, err)
		}
		// Name field: 'Q' then "$#@" then blanks.
		q, err := codec.Encode("Q")
		if err != nil {
			t.Fatal(err)
		}
		want := append(append([]byte{}, q...), tc.variant...)
		want = append(want, bytes.Repeat([]byte{0x40}, 10-len(want))...)
		if !bytes.Equal(obj.Queue[:], want) {
			t.Errorf("CCSID %d queue name = % x, want % x", tc.id, obj.Queue[:], want)
		}
		// Library "QGPL" encodes through the same codec, padded with blanks.
		lib, err := codec.Encode("QGPL")
		if err != nil {
			t.Fatal(err)
		}
		wantLib := append(append([]byte{}, lib...), bytes.Repeat([]byte{0x40}, 10-len(lib))...)
		if !bytes.Equal(obj.Library[:], wantLib) {
			t.Errorf("CCSID %d library = % x, want % x", tc.id, obj.Library[:], wantLib)
		}
	}
}
