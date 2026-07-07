package msdnsp

// ZoneUpdate is the value carried by a DSPROPERTY_ZONE_ALLOW_UPDATE property (fAllowUpdate). It
// identifies the dynamic-update policy of a zone.
//
// A value of ZONE_UPDATE_UNSECURE means the zone accepts unauthenticated dynamic updates, which
// allows any host to create or overwrite records in the zone.
//
// Source: [MS-DNSP] fAllowUpdate, referenced from DNS_RPC_ZONE_INFO (section 2.2.5.2.4.1)
type ZoneUpdate uint32

// ZONE_UPDATE_* constants.
const (
	ZONE_UPDATE_OFF      ZoneUpdate = 0x00000000 // Dynamic updates are not allowed.
	ZONE_UPDATE_UNSECURE ZoneUpdate = 0x00000001 // Both secure and unsecure dynamic updates are allowed.
	ZONE_UPDATE_SECURE   ZoneUpdate = 0x00000002 // Only secure dynamic updates are allowed.
)

var zoneUpdateNames = map[ZoneUpdate]string{
	ZONE_UPDATE_OFF:      "ZONE_UPDATE_OFF",
	ZONE_UPDATE_UNSECURE: "ZONE_UPDATE_UNSECURE",
	ZONE_UPDATE_SECURE:   "ZONE_UPDATE_SECURE",
}

// String returns the constant name of the update policy, or a hexadecimal fallback of the form
// "ZoneUpdate(0x12345678)" for undefined values.
func (u ZoneUpdate) String() string {
	if name, ok := zoneUpdateNames[u]; ok {
		return name
	}
	return "ZoneUpdate(0x" + hex32(uint32(u)) + ")"
}
