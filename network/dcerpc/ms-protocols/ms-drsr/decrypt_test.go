package msdrsr

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// Cross-implementation vector generated with impacket's deriveKey + the documented
// transport/RID layers (RID 500, NT hash of the empty password, a fixed 16-byte session
// key and salt). It validates transformKey, deriveKeysFromRID, removeDESLayer, and
// decryptReplicatedValue against the reference implementation, not just against itself.
func TestDeriveKeysFromRIDVector(t *testing.T) {
	k1, k2 := deriveKeysFromRID(500)
	if got := hex.EncodeToString(k1[:]); got != "f40040000ea00400" {
		t.Errorf("key1 = %s, want f40040000ea00400", got)
	}
	if got := hex.EncodeToString(k2[:]); got != "007a00200006d002" {
		t.Errorf("key2 = %s, want 007a00200006d002", got)
	}
}

func TestDecryptHashVector(t *testing.T) {
	sessionKey, _ := hex.DecodeString("30313233343536373839616263646566")
	payload, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0fea8f7d964f4156185b97105fd7f8d708339ff210")
	want, _ := hex.DecodeString("31d6cfe0d16ae931b73c59d7e0c089c0")

	got, err := decryptHash(sessionKey, payload, 500)
	if err != nil {
		t.Fatalf("decryptHash: %v", err)
	}
	if !bytes.Equal(got[:], want) {
		t.Errorf("NT hash = %x, want %x", got, want)
	}
}

// TestBerEncodeOIDContent checks the BER content octets of the Microsoft attribute OID
// for unicodePwd (1.2.840.113556.1.4.90).
func TestBerEncodeOIDContent(t *testing.T) {
	got, err := berEncodeOIDContent(oidUnicodePwd)
	if err != nil {
		t.Fatalf("berEncodeOIDContent: %v", err)
	}
	want, _ := hex.DecodeString("2a864886f71401045a")
	if !bytes.Equal(got, want) {
		t.Errorf("BER = %x, want %x", got, want)
	}
}

// TestAttidForOID checks the MakeAttid computation: with the 1.2.840.113556.1.4 prefix at
// table index 9 (the usual AD layout), unicodePwd resolves to 0x9005A.
func TestAttidForOID(t *testing.T) {
	msPrefix, _ := hex.DecodeString("2a864886f7140104") // 1.2.840.113556.1.4
	pt := drsrtypes.SCHEMA_PREFIX_TABLE{
		PrefixCount: 1,
		PPrefixEntry: []drsrtypes.PrefixTableEntry{
			{Ndx: 9, Prefix: drsrtypes.OID_t{Length: uint32(len(msPrefix)), Elements: msPrefix}},
		},
	}
	got, ok := attidForOID(pt, oidUnicodePwd)
	if !ok {
		t.Fatal("attidForOID: prefix not found")
	}
	if got != 0x9005A {
		t.Errorf("attid = 0x%X, want 0x9005A", got)
	}
	if _, ok := attidForOID(pt, oidUserPrincipalName); ok {
		// userPrincipalName shares the same prefix, so it should resolve too; its arc 656
		// is >= 128, exercising the 2-byte-prefix branch.
		got2, _ := attidForOID(pt, oidUserPrincipalName)
		if got2 != 0x90290 {
			t.Errorf("userPrincipalName attid = 0x%X, want 0x90290", got2)
		}
	}
}

// TestRidFromSID extracts the RID (last sub-authority) from a binary SID
// S-1-5-21-1-2-3-500.
func TestRidFromSID(t *testing.T) {
	sid := []byte{0x01, 0x05, 0, 0, 0, 0, 0, 0x05} // rev 1, 5 sub-authorities, authority 5
	for _, sub := range []uint32{21, 1, 2, 3, 500} {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], sub)
		sid = append(sid, b[:]...)
	}
	rid, err := ridFromSID(sid)
	if err != nil {
		t.Fatalf("ridFromSID: %v", err)
	}
	if rid != 500 {
		t.Errorf("rid = %d, want 500", rid)
	}
}
