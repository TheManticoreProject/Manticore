// Package msdnsp holds the NDR wire structures of the DNS Server Management Protocol
// ([MS-DNSP]), shared by the DnsServer interface (UUID 50abc2a4-574d-40b3-9d66-ee4fd5fba076
// version 5.0). Method stubs live under the interface tree; this package holds only the
// data types, one per file, with round-trip tests in structures_test.go.
//
// These types were generated from the [MS-DNSP] Appendix A IDL and reconciled by hand
// against the dcerpc-interface-structure skill's NDR rules. The Go round-trip tests are
// the acceptance gate available without a live DNS server; the following spots are
// modeled per the spec/skill but NOT yet validated against Windows on the wire:
//
//   - DNSSRV_RPC_UNION: a switch_type(DWORD) union whose discriminant is carried inline
//     (the Tag field) in addition to the method's dwTypeId argument, per Manticore's
//     switch_is convention. Each arm is a [unique] pointer to its payload struct.
//   - ASCII vs. wide strings: [string] char* fields are modeled as *ndr.STR (UTF-8) and
//     [string] wchar_t*/LPWSTR as *ndr.WSTR (UTF-16); the choice is per-field from the IDL.
//   - The [1]-sentinel enumeration lists (DNS_RPC_ENUM_ZONE_SCOPE_LIST, DNS_RPC_ENUM_SCOPE_LIST,
//     DNS_RPC_ENUM_VIRTUALIZATION_INSTANCE_LIST, DNS_RPC_ZONE_DNSSEC_SETTINGS.pZoneSkdArray)
//     keep the IDL's literal [1] declarator; on the wire the server sends dwXxxCount entries.
//   - The DNS-record data payloads inside DNS_RPC_RECORD.Buffer are an opaque byte blob here.
package msdnsp
