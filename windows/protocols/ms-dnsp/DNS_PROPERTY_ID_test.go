package msdnsp_test

import (
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestPropertyIdString verifies known ids render as their constant names and unknown ids fall
// back to a hex form.
func TestPropertyIdString(t *testing.T) {
	cases := []struct {
		id   msdnsp.PropertyId
		want string
	}{
		{msdnsp.DSPROPERTY_ZONE_TYPE, "DSPROPERTY_ZONE_TYPE"},
		{msdnsp.DSPROPERTY_ZONE_ALLOW_UPDATE, "DSPROPERTY_ZONE_ALLOW_UPDATE"},
		{msdnsp.DSPROPERTY_ZONE_NODE_DBFLAGS, "DSPROPERTY_ZONE_NODE_DBFLAGS"},
		{msdnsp.PropertyId(0xDEADBEEF), "PropertyId(0xdeadbeef)"},
	}
	for _, c := range cases {
		if got := c.id.String(); got != c.want {
			t.Errorf("PropertyId(0x%x).String() = %q; want %q", uint32(c.id), got, c.want)
		}
	}
}

// TestZoneTypeString verifies ZoneType rendering.
func TestZoneTypeString(t *testing.T) {
	if got := msdnsp.DNS_ZONE_TYPE_PRIMARY.String(); got != "DNS_ZONE_TYPE_PRIMARY" {
		t.Errorf("ZoneType.String() = %q; want DNS_ZONE_TYPE_PRIMARY", got)
	}
	if got := msdnsp.ZoneType(0x99).String(); got != "ZoneType(0x00000099)" {
		t.Errorf("unknown ZoneType.String() = %q; want hex fallback", got)
	}
}

// TestZoneUpdateString verifies ZoneUpdate rendering.
func TestZoneUpdateString(t *testing.T) {
	if got := msdnsp.ZONE_UPDATE_UNSECURE.String(); got != "ZONE_UPDATE_UNSECURE" {
		t.Errorf("ZoneUpdate.String() = %q; want ZONE_UPDATE_UNSECURE", got)
	}
	if got := msdnsp.ZoneUpdate(0x7).String(); got != "ZoneUpdate(0x00000007)" {
		t.Errorf("unknown ZoneUpdate.String() = %q; want hex fallback", got)
	}
}
