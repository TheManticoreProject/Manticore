package pac

import (
	"testing"
	"time"
)

// TestLogonInfoAccessors forges a PAC, parses it back, decodes the logon-info
// buffer through PAC.LogonInfo, and asserts the typed accessors reconstruct the
// user, group, and SID information (an encode→decode symmetry check that also
// covers the SID-composition and string helpers).
func TestLogonInfoAccessors(t *testing.T) {
	key := make([]byte, 16)
	info := sampleInfo(t)

	p, err := Forge(info, "Administrator", time.Unix(1700000000, 0), 23)
	if err != nil {
		t.Fatalf("Forge: %v", err)
	}
	signed, err := p.Sign(key, key)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parsed, err := Parse(signed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	li, err := parsed.LogonInfo()
	if err != nil {
		t.Fatalf("LogonInfo: %v", err)
	}

	if li.UserName() != "Administrator" {
		t.Errorf("UserName = %q, want Administrator", li.UserName())
	}
	if li.LogonDomainNameString() != "CORP" {
		t.Errorf("LogonDomainNameString = %q, want CORP", li.LogonDomainNameString())
	}
	if li.LogonServerName() != "DC01" {
		t.Errorf("LogonServerName = %q, want DC01", li.LogonServerName())
	}
	if li.UserRID() != 500 {
		t.Errorf("UserRID = %d, want 500", li.UserRID())
	}
	if li.PrimaryGroupRID() != 513 {
		t.Errorf("PrimaryGroupRID = %d, want 513", li.PrimaryGroupRID())
	}
	if li.AccountControl() != 0x00010200 {
		t.Errorf("AccountControl = %#x, want 0x00010200", li.AccountControl())
	}

	domain := "S-1-5-21-1111111111-2222222222-3333333333"
	if li.DomainSID() == nil || li.DomainSID().String() != domain {
		t.Errorf("DomainSID = %v, want %s", li.DomainSID(), domain)
	}
	if got, want := li.UserSID(), domain+"-500"; got != want {
		t.Errorf("UserSID = %q, want %q", got, want)
	}
	if got, want := li.PrimaryGroupSID(), domain+"-513"; got != want {
		t.Errorf("PrimaryGroupSID = %q, want %q", got, want)
	}

	rids := li.GroupRIDs()
	if len(rids) != 3 || rids[0] != 513 || rids[1] != 512 || rids[2] != 519 {
		t.Errorf("GroupRIDs = %v, want [513 512 519]", rids)
	}
	gsids := li.GroupSIDs()
	if len(gsids) != 3 || gsids[1] != domain+"-512" {
		t.Errorf("GroupSIDs = %v", gsids)
	}
	esids := li.ExtraSIDs()
	if len(esids) != 1 || esids[0] != domain+"-519" {
		t.Errorf("ExtraSIDs = %v, want [%s-519]", esids, domain)
	}
}

// TestFileTimeConversions checks the FILETIME helpers: the "never" sentinel and a
// zero value both map to the zero time, and a real timestamp round-trips through
// FileTimeFromTime.
func TestFileTimeConversions(t *testing.T) {
	if !NeverExpireFileTime().IsNever() {
		t.Error("NeverExpireFileTime().IsNever() = false")
	}
	if !NeverExpireFileTime().Time().IsZero() {
		t.Error("never FILETIME did not map to zero time")
	}
	if !(FILETIME{}).Time().IsZero() {
		t.Error("zero FILETIME did not map to zero time")
	}
	want := time.Unix(1700000000, 0).UTC()
	if got := FileTimeFromTime(want).Time(); !got.Equal(want) {
		t.Errorf("FILETIME round-trip = %v, want %v", got, want)
	}
}

// TestLogonInfoMissingBuffer ensures LogonInfo reports an error when the PAC has
// no logon-info buffer.
func TestLogonInfoMissingBuffer(t *testing.T) {
	p := &PAC{Buffers: []Buffer{{Type: BufferClientInfo, Data: []byte("x")}}}
	if _, err := p.LogonInfo(); err == nil {
		t.Error("expected error for PAC without a logon-info buffer")
	}
}
