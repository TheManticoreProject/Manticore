package dnsproperty

// PropertyId is the 32-bit value stored in the Id field of a DNS_PROPERTY (section 2.3.2.1). It
// specifies the type of data carried in the property's Data field.
//
// Source: [MS-DNSP] Property Id (section 2.3.2.1.1)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/3af63871-0cc4-4179-916c-5caade55a8f3
type PropertyId uint32

// DSPROPERTY_ZONE_* constants, per [MS-DNSP] section 2.3.2.1.1.
const (
	DSPROPERTY_ZONE_TYPE                  PropertyId = 0x00000001 // Zone type (dwZoneType). See ZoneType.
	DSPROPERTY_ZONE_ALLOW_UPDATE          PropertyId = 0x00000002 // Dynamic-update policy (fAllowUpdate). See ZoneUpdate.
	DSPROPERTY_ZONE_SECURE_TIME           PropertyId = 0x00000008 // Time at which the zone became secure.
	DSPROPERTY_ZONE_NOREFRESH_INTERVAL    PropertyId = 0x00000010 // No-refresh interval, in hours.
	DSPROPERTY_ZONE_SCAVENGING_SERVERS    PropertyId = 0x00000011 // Scavenging servers, as an IP4_ARRAY.
	DSPROPERTY_ZONE_AGING_ENABLED_TIME    PropertyId = 0x00000012 // Time interval before the next scavenging cycle.
	DSPROPERTY_ZONE_REFRESH_INTERVAL      PropertyId = 0x00000020 // Refresh interval, in hours.
	DSPROPERTY_ZONE_AGING_STATE           PropertyId = 0x00000040 // Whether aging is enabled (fAging).
	DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME PropertyId = 0x00000080 // Name of the server that deleted the zone (Unicode string).
	DSPROPERTY_ZONE_MASTER_SERVERS        PropertyId = 0x00000081 // Zone-transfer master servers, as an IP4_ARRAY.
	DSPROPERTY_ZONE_AUTO_NS_SERVERS       PropertyId = 0x00000082 // Servers that may auto-create a delegation, as an IP4_ARRAY.
	DSPROPERTY_ZONE_DCPROMO_CONVERT       PropertyId = 0x00000083 // State of DcPromo zone conversion. See DcPromo Flag (section 2.3.2.1.2).
	DSPROPERTY_ZONE_SCAVENGING_SERVERS_DA PropertyId = 0x00000090 // Scavenging servers, as a DNS_ADDR_ARRAY.
	DSPROPERTY_ZONE_MASTER_SERVERS_DA     PropertyId = 0x00000091 // Zone-transfer master servers, as a DNS_ADDR_ARRAY.
	DSPROPERTY_ZONE_AUTO_NS_SERVERS_DA    PropertyId = 0x00000092 // Auto-created-NS servers, as a DNS_ADDR_ARRAY.
	DSPROPERTY_ZONE_NODE_DBFLAGS          PropertyId = 0x00000100 // Node database flags. See DNS_RPC_NODE_FLAGS (section 2.2.2.1.2).
)

// propertyIdNames maps the defined PropertyId values to their [MS-DNSP] constant names for
// String rendering.
var propertyIdNames = map[PropertyId]string{
	DSPROPERTY_ZONE_TYPE:                  "DSPROPERTY_ZONE_TYPE",
	DSPROPERTY_ZONE_ALLOW_UPDATE:          "DSPROPERTY_ZONE_ALLOW_UPDATE",
	DSPROPERTY_ZONE_SECURE_TIME:           "DSPROPERTY_ZONE_SECURE_TIME",
	DSPROPERTY_ZONE_NOREFRESH_INTERVAL:    "DSPROPERTY_ZONE_NOREFRESH_INTERVAL",
	DSPROPERTY_ZONE_SCAVENGING_SERVERS:    "DSPROPERTY_ZONE_SCAVENGING_SERVERS",
	DSPROPERTY_ZONE_AGING_ENABLED_TIME:    "DSPROPERTY_ZONE_AGING_ENABLED_TIME",
	DSPROPERTY_ZONE_REFRESH_INTERVAL:      "DSPROPERTY_ZONE_REFRESH_INTERVAL",
	DSPROPERTY_ZONE_AGING_STATE:           "DSPROPERTY_ZONE_AGING_STATE",
	DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME: "DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME",
	DSPROPERTY_ZONE_MASTER_SERVERS:        "DSPROPERTY_ZONE_MASTER_SERVERS",
	DSPROPERTY_ZONE_AUTO_NS_SERVERS:       "DSPROPERTY_ZONE_AUTO_NS_SERVERS",
	DSPROPERTY_ZONE_DCPROMO_CONVERT:       "DSPROPERTY_ZONE_DCPROMO_CONVERT",
	DSPROPERTY_ZONE_SCAVENGING_SERVERS_DA: "DSPROPERTY_ZONE_SCAVENGING_SERVERS_DA",
	DSPROPERTY_ZONE_MASTER_SERVERS_DA:     "DSPROPERTY_ZONE_MASTER_SERVERS_DA",
	DSPROPERTY_ZONE_AUTO_NS_SERVERS_DA:    "DSPROPERTY_ZONE_AUTO_NS_SERVERS_DA",
	DSPROPERTY_ZONE_NODE_DBFLAGS:          "DSPROPERTY_ZONE_NODE_DBFLAGS",
}

// String returns the [MS-DNSP] constant name of the property id (for example
// "DSPROPERTY_ZONE_TYPE"), or a hexadecimal fallback of the form "PropertyId(0x12345678)" for
// values not defined in the specification's table.
func (id PropertyId) String() string {
	if name, ok := propertyIdNames[id]; ok {
		return name
	}
	const hexDigits = "0123456789abcdef"
	buf := []byte("PropertyId(0x00000000)")
	for i := 0; i < 8; i++ {
		buf[len(buf)-2-i] = hexDigits[(id>>(4*i))&0xF]
	}
	return string(buf)
}
