// Package dnsrecord implements the byte-level wire structures of the DNS resource
// records that Active Directory stores in the dnsRecord LDAP attribute, as specified
// by the Domain Name Service (DNS) Server Management Protocol [MS-DNSP].
//
// Unlike the NDR-marshaled RPC structures under windows/protocols/ms-dnsp (which are
// exchanged over the DnsServer RPC interface), the structures here use the packed,
// mixed-endian on-directory format defined for the dnsRecord attribute in [MS-DNSP]
// section 2.3.2.2. This is the format read and written by LDAP tooling that manipulates
// AD-integrated DNS zones directly (for example dnstool.py from krbrelayx).
//
// The outer container is DNS_RECORD (section 2.3.2.2). Its Data field carries one of the
// type-specific record-data payloads from DNS_RPC_RECORD_DATA (section 2.2.2.2.4), selected
// by the Type field (a DNS_RECORD_TYPE, section 2.2.2.1.1). The payloads implemented here
// are the ones commonly manipulated over LDAP:
//
//   - DNS_RPC_RECORD_A               (section 2.2.2.2.4.1)  - A
//   - DNS_RPC_RECORD_AAAA            (section 2.2.2.2.4.17) - AAAA
//   - DNS_RPC_RECORD_NODE_NAME       (section 2.2.2.2.4.2)  - NS, PTR, CNAME, DNAME, ...
//   - DNS_RPC_RECORD_SOA             (section 2.2.2.2.4.3)  - SOA
//   - DNS_RPC_RECORD_NAME_PREFERENCE (section 2.2.2.2.4.8)  - MX, AFSDB, RT
//   - DNS_RPC_RECORD_SRV             (section 2.2.2.2.4.7)  - SRV
//   - DNS_RPC_RECORD_TS             (section 2.2.2.2.4.23) - tombstone
//
// Name fields inside these payloads use DNS_COUNT_NAME (section 2.2.2.2.2), the LDAP
// on-directory form of a name, rather than the RPC-wire DNS_RPC_NAME form: when dnsRecord
// values are written over LDAP each name is converted from DNS_RPC_NAME to DNS_COUNT_NAME,
// and converted back when read.
//
// Every structure exposes Marshal() ([]byte, error) and Unmarshal([]byte) (int, error),
// where Unmarshal returns the number of bytes consumed. The multi-byte integer fields are
// little-endian except where [MS-DNSP] mandates network byte order (big-endian): the
// DNS_RECORD.TtlSeconds field, and the numeric fields of the SOA, SRV, and NAME_PREFERENCE
// payloads.
//
// Reference: [MS-DNSP] dnsRecord,
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/6912b338-5472-4f59-b912-0edb536b6ed8
package dnsrecord
