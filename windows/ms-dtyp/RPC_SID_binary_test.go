package msdtyp

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// TestRPCSIDBinaryCodec checks the packed binary SID form against known
// encodings ([MS-DTYP] 2.4.2.2) in both directions.
func TestRPCSIDBinaryCodec(t *testing.T) {
	for _, tc := range []struct{ text, hexBytes string }{
		{"S-1-5-32-544", "01020000000000052000000020020000"},
		{"S-1-1-0", "010100000000000100000000"},
		{"S-1-5-21-1004336348-1177238915-682003330-512", "010500000000000515000000dcf4dc3b833d2b46828ba62800020000"},
		{"S-1-0x123456789abc-7", "0101123456789abc07000000"},
		{"S-1-5", "0100000000000005"},
	} {
		want, _ := hex.DecodeString(tc.hexBytes)
		sid, err := ParseSID(tc.text)
		if err != nil {
			t.Fatalf("ParseSID(%s): %v", tc.text, err)
		}
		got, err := sid.Marshal()
		if err != nil {
			t.Fatalf("Marshal(%s): %v", tc.text, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Marshal(%s) = %x, want %x", tc.text, got, want)
		}

		var back RPC_SID
		n, err := back.Unmarshal(append(append([]byte(nil), want...), 0xEE, 0xEE))
		if err != nil || n != len(want) {
			t.Fatalf("Unmarshal(%x): n=%d err=%v", want, n, err)
		}
		if back.String() != tc.text {
			t.Errorf("Unmarshal(%x) = %s, want %s", want, back.String(), tc.text)
		}
		if int(back.SubAuthorityCount) != len(back.SubAuthority) {
			t.Errorf("SubAuthorityCount %d for %d sub-authorities", back.SubAuthorityCount, len(back.SubAuthority))
		}
	}
}

// TestRPCSIDBinaryRejectsMalformed checks truncated input and an oversized count
// are refused, on both sides.
func TestRPCSIDBinaryRejectsMalformed(t *testing.T) {
	for _, data := range [][]byte{
		nil,
		{1, 0, 0, 0, 0, 0, 0},                   // 7 bytes
		{1, 2, 0, 0, 0, 0, 0, 5, 0x20, 0, 0, 0}, // count 2, one sub-authority
		append([]byte{1, 16, 0, 0, 0, 0, 0, 5}, make([]byte, 64)...), // count 16
	} {
		var s RPC_SID
		if _, err := s.Unmarshal(data); err == nil {
			t.Errorf("Unmarshal(%x) accepted malformed input", data)
		}
	}
	if _, err := (RPC_SID{Revision: 1, SubAuthority: make([]uint32, 16)}).Marshal(); err == nil {
		t.Error("Marshal accepted 16 sub-authorities")
	}
}
