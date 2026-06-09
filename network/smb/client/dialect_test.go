package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

func TestSMB2DialectFor(t *testing.T) {
	cases := []struct {
		v    smb.SMBProtocolVersion
		want dialects.Dialect
		ok   bool
	}{
		{smb.SMB_VERSION_2_0, dialects.SMB2_DIALECT_2_0_2, true}, // family marker -> 2.0.2
		{smb.SMB_VERSION_2_0_2, dialects.SMB2_DIALECT_2_0_2, true},
		{smb.SMB_VERSION_2_1, dialects.SMB2_DIALECT_2_1_0, true},
		{smb.SMB_VERSION_3_0, dialects.SMB2_DIALECT_3_0_0, true},
		{smb.SMB_VERSION_3_0_2, dialects.SMB2_DIALECT_3_0_2, true},
		{smb.SMB_VERSION_3_1_1, dialects.SMB2_DIALECT_3_1_1, true},
		{smb.SMB_VERSION_1_0, 0, false}, // SMB1 has no SMB2 revision
	}
	for _, c := range cases {
		got, ok := smb2DialectFor(c.v)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("smb2DialectFor(%s) = (0x%04x, %v), want (0x%04x, %v)", c.v, uint16(got), ok, uint16(c.want), c.ok)
		}
	}
}

func TestVersionForSMB2Dialect_RoundTrip(t *testing.T) {
	// Every real SMB2 dialect must round-trip back to a concrete version.
	for _, d := range []dialects.Dialect{
		dialects.SMB2_DIALECT_2_0_2, dialects.SMB2_DIALECT_2_1_0,
		dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1,
	} {
		v, ok := versionForSMB2Dialect(d)
		if !ok {
			t.Errorf("versionForSMB2Dialect(0x%04x) ok=false, want true", uint16(d))
			continue
		}
		back, ok := smb2DialectFor(v)
		if !ok || back != d {
			t.Errorf("round-trip 0x%04x -> %s -> 0x%04x failed", uint16(d), v, uint16(back))
		}
	}
}

func TestVersionForSMB2Dialect_RejectsNonDialects(t *testing.T) {
	// The wildcard is not a concrete dialect.
	if _, ok := versionForSMB2Dialect(dialects.SMB2_DIALECT_WILDCARD); ok {
		t.Error("wildcard 0x02FF must not map to a concrete version")
	}
	// 0x0200 is the abstract family marker, not a real wire dialect revision.
	if _, ok := versionForSMB2Dialect(dialects.Dialect(0x0200)); ok {
		t.Error("0x0200 must not be accepted as a wire dialect")
	}
}

func TestDefaultPreference(t *testing.T) {
	pref := defaultPreference()
	if len(pref) == 0 {
		t.Fatal("defaultPreference() is empty")
	}
	for _, v := range pref {
		if !v.IsSupported() {
			t.Errorf("default preference contains unsupported version %s", v)
		}
	}
	// Must be ordered best-first: each entry strictly greater than the next.
	for i := 1; i < len(pref); i++ {
		if !(pref[i-1] > pref[i]) {
			t.Errorf("default preference not strictly descending at %d: %s !> %s", i, pref[i-1], pref[i])
		}
	}
}
