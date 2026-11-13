package ccsid_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
)

// TestObjectNameVariantCharactersEncodeWithJobCCSID pins how the three EBCDIC
// variant characters $, # and @ of the IBM i system-name subset are placed on the
// wire: through the job CCSID's codec, so each travels as that CCSID's own code
// point. This is correct — it matches the reference client, IBM Toolbox for Java
// (JTOpen), whose RemoteCommandImplRemote encodes object, library, program and
// command names with the command-server job CCSID converter, not a fixed CCSID 37.
// The expected bytes are the per-CCSID code points for "$#@"; a table regression
// that moved any of them would silently corrupt names sent to a job running that
// CCSID. (Settles #1911: no CCSID-37 override is applied to object names.)
func TestObjectNameVariantCharactersEncodeWithJobCCSID(t *testing.T) {
	cases := []struct {
		id   ccsid.CCSID
		want []byte
	}{
		{37, []byte{0x5B, 0x7B, 0x7C}},
		{500, []byte{0x5B, 0x7B, 0x7C}},
		{1140, []byte{0x5B, 0x7B, 0x7C}},
		{273, []byte{0x5B, 0x7B, 0xB5}},
		{277, []byte{0x67, 0x4A, 0x80}},
		{278, []byte{0x67, 0x63, 0xEC}},
		{280, []byte{0x5B, 0xB1, 0xB5}},
		{284, []byte{0x5B, 0x69, 0x7C}},
		{285, []byte{0x4A, 0x7B, 0x7C}},
		{297, []byte{0x5B, 0xB1, 0x44}},
		{871, []byte{0x5B, 0x7B, 0xAC}},
	}
	for _, tc := range cases {
		c, err := ccsid.Lookup(tc.id)
		if err != nil {
			t.Fatalf("Lookup(%d): %v", tc.id, err)
		}
		got, err := c.Encode("$#@")
		if err != nil {
			t.Fatalf("CCSID %d Encode(%q): %v", tc.id, "$#@", err)
		}
		if !bytes.Equal(got, tc.want) {
			t.Errorf("CCSID %d Encode(\"$#@\") = % X, want % X", tc.id, got, tc.want)
		}
	}
}
