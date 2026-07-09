package keytab

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

func sampleKeytab() *Keytab {
	kt := New()
	admin := Principal{NameType: iana.NameTypePrincipal, Realm: "CORP.LOCAL", Components: []string{"Administrator"}}
	// AES256, AES128 and RC4 keys for the same principal at kvno 3.
	kt.Add(admin, iana.ETypeAES256CTSHMACSHA196, bytes.Repeat([]byte{0xAB}, 32), 3)
	kt.Add(admin, iana.ETypeAES128CTSHMACSHA196, bytes.Repeat([]byte{0xCD}, 16), 3)
	kt.Add(admin, iana.ETypeRC4HMAC, bytes.Repeat([]byte{0xEF}, 16), 3)
	// A service principal with two components and a large kvno forcing the
	// 32-bit vno trailer.
	svc := Principal{NameType: iana.NameTypeSRVInst, Realm: "CORP.LOCAL", Components: []string{"HTTP", "web.corp.local"}}
	kt.Add(svc, iana.ETypeAES256CTSHMACSHA196, bytes.Repeat([]byte{0x11}, 32), 300)
	return kt
}

func TestKeytabRoundtrip(t *testing.T) {
	orig := sampleKeytab()
	b, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("round-trip mismatch:\n orig=%+v\n got =%+v", orig, got)
	}
	// Marshaling the parsed keytab must reproduce the identical bytes.
	b2, err := got.Marshal()
	if err != nil {
		t.Fatalf("re-Marshal: %v", err)
	}
	if !bytes.Equal(b, b2) {
		t.Fatalf("re-marshaled bytes differ from original")
	}
}

func TestKeytabKvnoTrailer(t *testing.T) {
	kt := sampleKeytab()
	b, _ := kt.Marshal()
	got, err := Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// The kvno-300 entry must carry the 32-bit vno and report it via Kvno().
	e := got.Select("HTTP/web.corp.local@CORP.LOCAL", 0, -1)
	if e == nil {
		t.Fatal("service entry not found")
	}
	if e.KVNO != 300 {
		t.Fatalf("expected 32-bit KVNO 300, got %d", e.KVNO)
	}
	if e.Kvno() != 300 {
		t.Fatalf("Kvno() = %d, want 300", e.Kvno())
	}
	// A small kvno stays in the 8-bit field only.
	admin := got.Select("Administrator@CORP.LOCAL", iana.ETypeRC4HMAC, -1)
	if admin == nil || admin.KVNO != 0 || admin.KVNO8 != 3 || admin.Kvno() != 3 {
		t.Fatalf("admin kvno fields wrong: %+v", admin)
	}
}

func TestKeytabSelect(t *testing.T) {
	kt := sampleKeytab()

	// No etype filter -> strongest (AES256) wins for the Administrator.
	best := kt.Select("Administrator@CORP.LOCAL", 0, -1)
	if best == nil || int(best.EType) != iana.ETypeAES256CTSHMACSHA196 {
		t.Fatalf("expected AES256 as strongest, got %+v", best)
	}
	// Explicit etype filter selects that etype.
	rc4 := kt.Select("Administrator@CORP.LOCAL", iana.ETypeRC4HMAC, -1)
	if rc4 == nil || int(rc4.EType) != iana.ETypeRC4HMAC {
		t.Fatalf("expected RC4, got %+v", rc4)
	}
	// Realm-less query matches, case-insensitively on realm.
	if kt.Select("Administrator", 0, -1) == nil {
		t.Fatal("realm-less query should match")
	}
	if kt.Select("Administrator@corp.local", iana.ETypeAES128CTSHMACSHA196, -1) == nil {
		t.Fatal("case-insensitive realm match failed")
	}
	// Find with etype+kvno filters.
	if hits := kt.Find("Administrator@CORP.LOCAL", 0, 3); len(hits) != 3 {
		t.Fatalf("expected 3 entries at kvno 3, got %d", len(hits))
	}
	// Unknown principal / etype -> no match.
	if kt.Select("nobody@CORP.LOCAL", 0, -1) != nil {
		t.Fatal("unexpected match for unknown principal")
	}
	if kt.Select("Administrator@CORP.LOCAL", iana.ETypeDES3CBCSHA1KD, -1) != nil {
		t.Fatal("unexpected match for absent etype")
	}
}

// TestKeytabKnownVector parses a hand-assembled versioned-format-2 keytab and
// verifies each decoded field, then confirms Marshal reproduces the exact bytes.
func TestKeytabKnownVector(t *testing.T) {
	vec := []byte{
		0x05, 0x02, // file_format_version 0x0502
		0x00, 0x00, 0x00, 0x19, // entry size = 25 bytes
		0x00, 0x01, // num_components = 1
		0x00, 0x01, 0x52, // realm "R"
		0x00, 0x01, 0x4E, // component "N"
		0x00, 0x00, 0x00, 0x01, // name_type = NT-PRINCIPAL
		0x00, 0x00, 0x00, 0x00, // timestamp = 0
		0x02,       // vno8 = 2
		0x00, 0x17, // enctype = 23 (RC4-HMAC)
		0x00, 0x04, 0xDE, 0xAD, 0xBE, 0xEF, // key (4 bytes)
	}

	kt, err := Unmarshal(vec)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(kt.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(kt.Entries))
	}
	e := kt.Entries[0]
	if e.Principal.Realm != "R" ||
		!reflect.DeepEqual(e.Principal.Components, []string{"N"}) ||
		e.Principal.NameType != iana.NameTypePrincipal {
		t.Fatalf("principal wrong: %+v", e.Principal)
	}
	if e.Timestamp != 0 || e.KVNO8 != 2 || e.KVNO != 0 || e.Kvno() != 2 {
		t.Fatalf("version/time fields wrong: %+v", e)
	}
	if int(e.EType) != iana.ETypeRC4HMAC || !bytes.Equal(e.Key, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Fatalf("key fields wrong: %+v", e)
	}

	out, err := kt.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(out, vec) {
		t.Fatalf("Marshal did not reproduce known vector:\n got=% x\n want=% x", out, vec)
	}
}

// TestKeytabHole verifies that a negative record length (a deleted entry / hole)
// is skipped and the following entry is still parsed.
func TestKeytabHole(t *testing.T) {
	kt := New()
	kt.Add(Principal{Realm: "R", Components: []string{"N"}}, iana.ETypeRC4HMAC, []byte{1, 2, 3, 4}, 1)
	b, _ := kt.Marshal()

	// Splice a 4-byte hole (size = -4) right after the version tag, before the
	// real entry.
	hole := []byte{0xFF, 0xFF, 0xFF, 0xFC, 0xAA, 0xAA, 0xAA, 0xAA} // -4 then 4 junk bytes
	withHole := append([]byte{}, b[:2]...)
	withHole = append(withHole, hole...)
	withHole = append(withHole, b[2:]...)

	got, err := Unmarshal(withHole)
	if err != nil {
		t.Fatalf("Unmarshal with hole: %v", err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Principal.Realm != "R" {
		t.Fatalf("hole not skipped correctly: %+v", got.Entries)
	}
}

func TestKeytabSaveLoad(t *testing.T) {
	kt := sampleKeytab()
	path := filepath.Join(t.TempDir(), "test.keytab")
	if err := kt.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(kt, got) {
		t.Fatalf("save/load mismatch")
	}
}

func TestKeytabUnmarshalErrors(t *testing.T) {
	if _, err := Unmarshal([]byte{0x05, 0x01}); err == nil {
		t.Fatal("expected error on version 1")
	}
	if _, err := Unmarshal([]byte{0x05}); err == nil {
		t.Fatal("expected error on truncated version")
	}
	// Size claims more bytes than remain.
	if _, err := Unmarshal([]byte{0x05, 0x02, 0x00, 0x00, 0x00, 0x40, 0x01}); err == nil {
		t.Fatal("expected truncation error")
	}
}
