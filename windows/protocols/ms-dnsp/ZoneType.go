package msdnsp

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

// hex32 renders v as 8 lowercase hexadecimal digits. It is shared by the ZoneType and
// ZoneUpdate String methods.
func hex32(v uint32) string {
	const hexDigits = "0123456789abcdef"
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[7-i] = hexDigits[(v>>(4*i))&0xF]
	}
	return string(buf)
}
