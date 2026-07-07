// Package dnsproperty implements the byte-level wire structure of the dnsProperty LDAP
// attribute that Active Directory stores on AD-integrated DNS zone objects, as specified by
// the Domain Name Service (DNS) Server Management Protocol [MS-DNSP].
//
// A dnsProperty value is the packed, on-directory form of a single zone property (zone type,
// dynamic-update policy, aging/scavenging settings, server lists, and so on). The dnsProperty
// attribute is multi-valued: each value carries exactly one property, so callers unmarshal one
// DNS_PROPERTY per attribute value.
//
// The container is DNS_PROPERTY (section 2.3.2.1): a fixed 20-byte header (DataLength,
// NameLength, Flag, Version, Id) followed by a variable-length Data field of DataLength bytes
// and a trailing 1-byte Name field that is not used. All header integers are little-endian.
// The Id field (a PropertyId, section 2.3.2.1.1) selects how the Data field is interpreted.
//
// The type-specific Data payloads are exposed through accessors on DNS_PROPERTY:
//
//   - AsUint32     - the 32-bit scalar properties (zone type, allow-update, intervals,
//     aging state, DcPromo flag, node DB flags, ...)
//   - AsIP4Array   - the IP4_ARRAY server lists (section 2.2.3.2.1: scavenging, master, and
//     auto-created-NS server lists)
//   - AsUTF16String - the DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME null-terminated Unicode string
//
// The two most security-relevant scalar properties additionally have typed enumerations with a
// String rendering: ZoneType (dwZoneType) and ZoneUpdate (fAllowUpdate).
//
// Reference: [MS-DNSP] dnsProperty,
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/445c7843-e4a1-4222-8c0f-630c230a4c80
package dnsproperty
