package dnsproperty

// ZoneType is the value carried by a DSPROPERTY_ZONE_TYPE property (dwZoneType). It identifies
// the kind of DNS zone.
//
// Source: [MS-DNSP] dwZoneType, referenced from DNS_RPC_ZONE_INFO (section 2.2.5.2.4.1)
type ZoneType uint32

// DNS_ZONE_TYPE_* constants.
const (
	DNS_ZONE_TYPE_CACHE           ZoneType = 0x00000000 // A cache (root hints) zone.
	DNS_ZONE_TYPE_PRIMARY         ZoneType = 0x00000001 // A primary zone.
	DNS_ZONE_TYPE_SECONDARY       ZoneType = 0x00000002 // A secondary zone.
	DNS_ZONE_TYPE_STUB            ZoneType = 0x00000003 // A stub zone.
	DNS_ZONE_TYPE_FORWARDER       ZoneType = 0x00000004 // A forwarder zone.
	DNS_ZONE_TYPE_SECONDARY_CACHE ZoneType = 0x00000005 // A secondary cache zone.
)

var zoneTypeNames = map[ZoneType]string{
	DNS_ZONE_TYPE_CACHE:           "DNS_ZONE_TYPE_CACHE",
	DNS_ZONE_TYPE_PRIMARY:         "DNS_ZONE_TYPE_PRIMARY",
	DNS_ZONE_TYPE_SECONDARY:       "DNS_ZONE_TYPE_SECONDARY",
	DNS_ZONE_TYPE_STUB:            "DNS_ZONE_TYPE_STUB",
	DNS_ZONE_TYPE_FORWARDER:       "DNS_ZONE_TYPE_FORWARDER",
	DNS_ZONE_TYPE_SECONDARY_CACHE: "DNS_ZONE_TYPE_SECONDARY_CACHE",
}

// String returns the constant name of the zone type, or a hexadecimal fallback of the form
// "ZoneType(0x12345678)" for undefined values.
func (t ZoneType) String() string {
	if name, ok := zoneTypeNames[t]; ok {
		return name
	}
	return "ZoneType(0x" + hex32(uint32(t)) + ")"
}

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

// hex32 renders v as 8 lowercase hexadecimal digits.
func hex32(v uint32) string {
	const hexDigits = "0123456789abcdef"
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[7-i] = hexDigits[(v>>(4*i))&0xF]
	}
	return string(buf)
}
