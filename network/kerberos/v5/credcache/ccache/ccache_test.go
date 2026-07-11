package ccache

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

func sampleCache() *CCache {
	return &CCache{
		DeltaTimeSecs: 42,
		DefaultPrincipal: Principal{
			NameType:   1,
			Realm:      "CORP.LOCAL",
			Components: []string{"alice"},
		},
		Credentials: []Credential{
			{
				Client:      Principal{NameType: 1, Realm: "CORP.LOCAL", Components: []string{"alice"}},
				Server:      Principal{NameType: 2, Realm: "CORP.LOCAL", Components: []string{"krbtgt", "CORP.LOCAL"}},
				Key:         Keyblock{EType: 18, KeyValue: bytes.Repeat([]byte{0xAB}, 32)},
				AuthTime:    1_700_000_000,
				StartTime:   1_700_000_000,
				EndTime:     1_700_036_000,
				RenewTill:   1_700_600_000,
				IsSKey:      false,
				TicketFlags: 0x50e00000, // forwardable+renewable+initial+pre-authent style bits
				Ticket:      bytes.Repeat([]byte{0x01, 0x02}, 20),
			},
			{
				Client:       Principal{NameType: 1, Realm: "CORP.LOCAL", Components: []string{"alice"}},
				Server:       Principal{NameType: 2, Realm: "CORP.LOCAL", Components: []string{"cifs", "dc01.corp.local"}},
				Key:          Keyblock{EType: 23, KeyValue: bytes.Repeat([]byte{0xCD}, 16)},
				EndTime:      1_700_036_000,
				TicketFlags:  0x40a00000,
				Addresses:    []Address{{AddrType: 2, Data: []byte{10, 0, 0, 1}}},
				AuthData:     []AuthData{{ADType: 1, Data: []byte{9, 9, 9}}},
				Ticket:       []byte("service-ticket-bytes"),
				SecondTicket: []byte{},
			},
		},
	}
}

func TestCCacheRoundtrip(t *testing.T) {
	orig := sampleCache()
	b, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Version tag must be 05 04.
	if b[0] != 0x05 || b[1] != 0x04 {
		t.Fatalf("version tag = % X, want 05 04", b[:2])
	}
	got, err := Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig.DefaultPrincipal, got.DefaultPrincipal) {
		t.Errorf("default principal mismatch:\n got %+v\n want %+v", got.DefaultPrincipal, orig.DefaultPrincipal)
	}
	if got.DeltaTimeSecs != orig.DeltaTimeSecs {
		t.Errorf("delta-time secs: got %d, want %d", got.DeltaTimeSecs, orig.DeltaTimeSecs)
	}
	if len(got.Credentials) != len(orig.Credentials) {
		t.Fatalf("credential count: got %d, want %d", len(got.Credentials), len(orig.Credentials))
	}
	for i := range orig.Credentials {
		// Normalize nil vs empty slice for the comparison of the second_ticket.
		if len(got.Credentials[i].SecondTicket) == 0 {
			got.Credentials[i].SecondTicket = orig.Credentials[i].SecondTicket
		}
		if !reflect.DeepEqual(got.Credentials[i], orig.Credentials[i]) {
			t.Errorf("credential[%d] mismatch:\n got  %+v\n want %+v", i, got.Credentials[i], orig.Credentials[i])
		}
	}
}

func TestCCacheSaveLoad(t *testing.T) {
	orig := sampleCache()
	path := filepath.Join(t.TempDir(), "krb5cc_test")
	if err := orig.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.DefaultPrincipal.Realm != "CORP.LOCAL" || len(got.Credentials) != 2 {
		t.Errorf("loaded cache wrong: %+v", got)
	}
	// The TGT's server must be krbtgt/CORP.LOCAL.
	tgt := got.Credentials[0].Server
	if len(tgt.Components) != 2 || tgt.Components[0] != "krbtgt" {
		t.Errorf("TGT server principal wrong: %+v", tgt)
	}
}

func TestCCacheRejectsWrongVersion(t *testing.T) {
	if _, err := Unmarshal([]byte{0x05, 0x03, 0x00, 0x00}); err == nil {
		t.Error("expected error for version 3 cache")
	}
	if _, err := Unmarshal([]byte{0x05}); err == nil {
		t.Error("expected error for truncated input")
	}
}

// TestCCacheRejectsHugeLength ensures a hostile variable-length field with a
// declared length that overflows the remaining buffer is rejected with a clean
// error instead of panicking in make([]byte, n). The 0x80000000 length also
// covers the negative-int path of need(): on a 32-bit build int(n) is negative,
// which the n < 0 guard rejects (mirroring the keytab reader); on a 64-bit build
// it is a large positive value caught by the pos+n > len(b) bounds check.
func TestCCacheRejectsHugeLength(t *testing.T) {
	// version 4 (05 04), empty header (00 00), then a principal whose realm
	// data field claims a length of 0x80000000 bytes.
	crafted := []byte{
		0x05, 0x04, // version 4
		0x00, 0x00, // header length 0
		0x00, 0x00, 0x00, 0x01, // principal name-type
		0x00, 0x00, 0x00, 0x00, // component count 0
		0x80, 0x00, 0x00, 0x00, // realm length 0x80000000 (hostile)
	}
	if _, err := Unmarshal(crafted); err == nil {
		t.Fatal("expected error for oversized length field, got nil")
	}

	// A plainly-too-large length that exceeds the buffer on any platform must
	// likewise return an error rather than reading out of bounds.
	crafted[12], crafted[13], crafted[14], crafted[15] = 0x00, 0x00, 0xFF, 0xFF
	if _, err := Unmarshal(crafted); err == nil {
		t.Fatal("expected error for length exceeding buffer, got nil")
	}
}

func TestCCacheEmptyHeaderNoDelta(t *testing.T) {
	cc := &CCache{DefaultPrincipal: Principal{NameType: 1, Realm: "R", Components: []string{"u"}}}
	b, err := cc.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// header length (bytes 2-3) must be 0 when no delta-time is set.
	if b[2] != 0 || b[3] != 0 {
		t.Errorf("expected empty header, got header len % X", b[2:4])
	}
	got, err := Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.DefaultPrincipal.Realm != "R" {
		t.Errorf("principal round-trip failed: %+v", got.DefaultPrincipal)
	}
}
