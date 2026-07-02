package msdrsr

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// msPrefixTable returns a prefix table with the 1.2.840.113556.1.4 attribute prefix at
// index 9, as a default DC reply carries it.
func msPrefixTable() drsrtypes.SCHEMA_PREFIX_TABLE {
	p, _ := hex.DecodeString("2a864886f7140104")
	return drsrtypes.SCHEMA_PREFIX_TABLE{
		PrefixCount:  1,
		PPrefixEntry: []drsrtypes.PrefixTableEntry{{Ndx: 9, Prefix: drsrtypes.OID_t{Length: uint32(len(p)), Elements: p}}},
	}
}

func sidWithRID(rid uint32) []byte {
	sid := []byte{0x01, 0x05, 0, 0, 0, 0, 0, 0x05}
	for _, sub := range []uint32{21, 1, 2, 3, rid} {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], sub)
		sid = append(sid, b[:]...)
	}
	return sid
}

// TestDecryptObjectSecrets wires the full extraction path: a synthetic replicated object
// carrying objectSid (RID 500) and the encrypted unicodePwd vector resolves to the NT
// hash of the empty password.
func TestDecryptObjectSecrets(t *testing.T) {
	sessionKey, _ := hex.DecodeString("30313233343536373839616263646566")
	encUnicodePwd, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0fea8f7d964f4156185b97105fd7f8d708339ff210")
	wantNT := "31d6cfe0d16ae931b73c59d7e0c089c0"

	obj := ReplicatedObject{
		DN: "CN=test,DC=lab,DC=local",
		Attributes: []ReplicatedAttribute{
			{AttrType: 0x90092, Values: [][]byte{sidWithRID(500)}}, // objectSid
			{AttrType: 0x9005A, Values: [][]byte{encUnicodePwd}},   // unicodePwd
		},
	}

	rid, err := ridFromSID(obj.Attributes[0].Values[0])
	if err != nil {
		t.Fatalf("ridFromSID: %v", err)
	}
	sec, err := decryptObjectSecrets(obj, msPrefixTable(), sessionKey, rid)
	if err != nil {
		t.Fatalf("decryptObjectSecrets: %v", err)
	}
	if !sec.HasNT {
		t.Fatal("NT hash not extracted")
	}
	if got := hex.EncodeToString(sec.NTHash[:]); got != wantNT {
		t.Errorf("NT hash = %s, want %s", got, wantNT)
	}
	if sec.RID != 500 {
		t.Errorf("RID = %d, want 500", sec.RID)
	}
}
