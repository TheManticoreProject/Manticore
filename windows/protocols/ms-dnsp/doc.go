// Package msdnsp holds the wire structures of the DNS Server Management Protocol ([MS-DNSP]),
// one type per file. It contains two distinct serialization families that share the one
// specification:
//
//  1. NDR/RPC structures - the types exchanged over the DnsServer RPC interface
//     (UUID 50abc2a4-574d-40b3-9d66-ee4fd5fba076 version 5.0). These were generated from the
//     [MS-DNSP] Appendix A IDL and reconciled by hand against the dcerpc-interface-structure
//     skill's NDR rules; they are marshaled with the ndr codec and round-tripped in
//     structures_test.go. Method stubs live under the interface tree, not here.
//
//  2. Packed LDAP on-directory structures - the byte-packed, mixed-endian formats that Active
//     Directory stores in DNS attributes and that LDAP tooling reads and writes directly. These
//     are NOT NDR: each exposes Marshal() ([]byte, error) and Unmarshal([]byte) (int, error)
//     and is verified by its own byte-level tests, so they are intentionally outside the
//     ndr.Marshal round-trip harness in structures_test.go. This family comprises:
//
//     - DNS_RECORD (section 2.3.2.2): the dnsRecord attribute container, with the record-data
//     payloads DNS_RPC_RECORD_A, _AAAA, _NODE_NAME, _SOA, _SRV, _NAME_PREFERENCE and _TS,
//     the DNS_COUNT_NAME name form (section 2.2.2.2.2), and the RecordType map
//     (DNS_RECORD_TYPE, section 2.2.2.1.1).
//     - DNS_PROPERTY (section 2.3.2.1): the dnsProperty attribute container, with the
//     PropertyId map (section 2.3.2.1.1) and the ZoneType / ZoneUpdate value enumerations.
//
// The two families do not overlap in names: the NDR DNS_RPC_RECORD (with its opaque Buffer
// field) is the RPC record wrapper, whereas the packed DNS_RECORD is the LDAP container and the
// DNS_RPC_RECORD_* types are its packed payloads. Header endianness in the packed family is
// little-endian except where [MS-DNSP] mandates network byte order (DNS_RECORD.TtlSeconds and
// the SOA/SRV/NAME_PREFERENCE numeric fields); DNS_RPC_RECORD_TS.EntombedTime is little-endian.
//
// The following NDR spots are modeled per the spec/skill but NOT yet validated against Windows
// on the wire:
//
//   - DNSSRV_RPC_UNION: a switch_type(DWORD) union whose discriminant is carried inline
//     (the Tag field) in addition to the method's dwTypeId argument, per Manticore's
//     switch_is convention. Each arm is a [unique] pointer to its payload struct.
//   - ASCII vs. wide strings: [string] char* fields are modeled as *ndr.STR (UTF-8) and
//     [string] wchar_t*/LPWSTR as *ndr.WSTR (UTF-16); the choice is per-field from the IDL.
//   - The [1]-sentinel enumeration lists (DNS_RPC_ENUM_ZONE_SCOPE_LIST, DNS_RPC_ENUM_SCOPE_LIST,
//     DNS_RPC_ENUM_VIRTUALIZATION_INSTANCE_LIST, DNS_RPC_ZONE_DNSSEC_SETTINGS.pZoneSkdArray)
//     keep the IDL's literal [1] declarator; on the wire the server sends dwXxxCount entries.
//   - The DNS-record data payloads inside the NDR DNS_RPC_RECORD.Buffer remain an opaque byte
//     blob; the packed LDAP DNS_RECORD family above is the decoded on-directory representation.
package msdnsp
