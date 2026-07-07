package dnsrecord

// RecordType is a 16-bit value that specifies a DNS resource record's type. It is stored
// in the Type field of a DNS_RECORD (section 2.3.2.2) and selects which DNS_RPC_RECORD_DATA
// (section 2.2.2.2.4) payload appears in the record's Data field.
//
// Source: [MS-DNSP] DNS_RECORD_TYPE (section 2.2.2.1.1)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/39b03b89-2264-4063-8198-d62f62a6441a
type RecordType uint16

// DNS_RECORD_TYPE constants, per [MS-DNSP] section 2.2.2.1.1. The values are the IANA DNS
// resource-record type codes, with the Microsoft-specific WINS/WINSR extensions in the
// private 0xFF00 range.
const (
	DNS_TYPE_ZERO       RecordType = 0x0000 // An empty record type (RFC1034 3.6, RFC1035 3.2.2).
	DNS_TYPE_A          RecordType = 0x0001 // An A record type, an IPv4 address (RFC1035 3.2.2).
	DNS_TYPE_NS         RecordType = 0x0002 // An authoritative name-server record type.
	DNS_TYPE_MD         RecordType = 0x0003 // A mail-destination record type.
	DNS_TYPE_MF         RecordType = 0x0004 // A mail-forwarder record type.
	DNS_TYPE_CNAME      RecordType = 0x0005 // The canonical name of a DNS alias.
	DNS_TYPE_SOA        RecordType = 0x0006 // A Start of Authority (SOA) record type.
	DNS_TYPE_MB         RecordType = 0x0007 // A mailbox record type.
	DNS_TYPE_MG         RecordType = 0x0008 // A mail group member record type.
	DNS_TYPE_MR         RecordType = 0x0009 // A mail-rename record type.
	DNS_TYPE_NULL       RecordType = 0x000A // A record type for completion queries.
	DNS_TYPE_WKS        RecordType = 0x000B // A well-known service record type.
	DNS_TYPE_PTR        RecordType = 0x000C // An FQDN pointer record type.
	DNS_TYPE_HINFO      RecordType = 0x000D // A host information record type.
	DNS_TYPE_MINFO      RecordType = 0x000E // A mailbox or mailing list information record type.
	DNS_TYPE_MX         RecordType = 0x000F // A mail-exchanger record type.
	DNS_TYPE_TXT        RecordType = 0x0010 // A text string record type.
	DNS_TYPE_RP         RecordType = 0x0011 // A responsible-person record type (RFC1183).
	DNS_TYPE_AFSDB      RecordType = 0x0012 // An AFS database location record type (RFC1183).
	DNS_TYPE_X25        RecordType = 0x0013 // An X25 PSDN address record type (RFC1183).
	DNS_TYPE_ISDN       RecordType = 0x0014 // An ISDN address record type (RFC1183).
	DNS_TYPE_RT         RecordType = 0x0015 // A route-through record type (RFC1183).
	DNS_TYPE_SIG        RecordType = 0x0018 // A cryptographic public key signature record type (RFC2931).
	DNS_TYPE_KEY        RecordType = 0x0019 // A DNSSEC public key record type (RFC2535).
	DNS_TYPE_AAAA       RecordType = 0x001C // An IPv6 address record type (RFC3596).
	DNS_TYPE_LOC        RecordType = 0x001D // A location information record type (RFC1876).
	DNS_TYPE_NXT        RecordType = 0x001E // A next-domain record type (RFC2065).
	DNS_TYPE_SRV        RecordType = 0x0021 // A server selection record type (RFC2782).
	DNS_TYPE_ATMA       RecordType = 0x0022 // An ATM address record type.
	DNS_TYPE_NAPTR      RecordType = 0x0023 // A NAPTR record type (RFC2915).
	DNS_TYPE_DNAME      RecordType = 0x0027 // A DNAME record type (RFC2672).
	DNS_TYPE_DS         RecordType = 0x002B // A DS record type (RFC4034).
	DNS_TYPE_RRSIG      RecordType = 0x002E // An RRSIG record type (RFC4034).
	DNS_TYPE_NSEC       RecordType = 0x002F // An NSEC record type (RFC4034).
	DNS_TYPE_DNSKEY     RecordType = 0x0030 // A DNSKEY record type (RFC4034).
	DNS_TYPE_DHCID      RecordType = 0x0031 // A DHCID record type (RFC4701).
	DNS_TYPE_NSEC3      RecordType = 0x0032 // An NSEC3 record type (RFC5155).
	DNS_TYPE_NSEC3PARAM RecordType = 0x0033 // An NSEC3PARAM record type (RFC5155).
	DNS_TYPE_TLSA       RecordType = 0x0034 // A TLSA record type (RFC6698).
	DNS_TYPE_ALL        RecordType = 0x00FF // A query-only type requesting all records.
	DNS_TYPE_WINS       RecordType = 0xFF01 // A WINS forward-lookup record type (MS-WINSRA).
	DNS_TYPE_WINSR      RecordType = 0xFF02 // A WINS reverse-lookup record type (MS-WINSRA).
)

// recordTypeNames maps the well-known DNS_RECORD_TYPE values to their [MS-DNSP] constant
// names for String rendering.
var recordTypeNames = map[RecordType]string{
	DNS_TYPE_ZERO:       "DNS_TYPE_ZERO",
	DNS_TYPE_A:          "DNS_TYPE_A",
	DNS_TYPE_NS:         "DNS_TYPE_NS",
	DNS_TYPE_MD:         "DNS_TYPE_MD",
	DNS_TYPE_MF:         "DNS_TYPE_MF",
	DNS_TYPE_CNAME:      "DNS_TYPE_CNAME",
	DNS_TYPE_SOA:        "DNS_TYPE_SOA",
	DNS_TYPE_MB:         "DNS_TYPE_MB",
	DNS_TYPE_MG:         "DNS_TYPE_MG",
	DNS_TYPE_MR:         "DNS_TYPE_MR",
	DNS_TYPE_NULL:       "DNS_TYPE_NULL",
	DNS_TYPE_WKS:        "DNS_TYPE_WKS",
	DNS_TYPE_PTR:        "DNS_TYPE_PTR",
	DNS_TYPE_HINFO:      "DNS_TYPE_HINFO",
	DNS_TYPE_MINFO:      "DNS_TYPE_MINFO",
	DNS_TYPE_MX:         "DNS_TYPE_MX",
	DNS_TYPE_TXT:        "DNS_TYPE_TXT",
	DNS_TYPE_RP:         "DNS_TYPE_RP",
	DNS_TYPE_AFSDB:      "DNS_TYPE_AFSDB",
	DNS_TYPE_X25:        "DNS_TYPE_X25",
	DNS_TYPE_ISDN:       "DNS_TYPE_ISDN",
	DNS_TYPE_RT:         "DNS_TYPE_RT",
	DNS_TYPE_SIG:        "DNS_TYPE_SIG",
	DNS_TYPE_KEY:        "DNS_TYPE_KEY",
	DNS_TYPE_AAAA:       "DNS_TYPE_AAAA",
	DNS_TYPE_LOC:        "DNS_TYPE_LOC",
	DNS_TYPE_NXT:        "DNS_TYPE_NXT",
	DNS_TYPE_SRV:        "DNS_TYPE_SRV",
	DNS_TYPE_ATMA:       "DNS_TYPE_ATMA",
	DNS_TYPE_NAPTR:      "DNS_TYPE_NAPTR",
	DNS_TYPE_DNAME:      "DNS_TYPE_DNAME",
	DNS_TYPE_DS:         "DNS_TYPE_DS",
	DNS_TYPE_RRSIG:      "DNS_TYPE_RRSIG",
	DNS_TYPE_NSEC:       "DNS_TYPE_NSEC",
	DNS_TYPE_DNSKEY:     "DNS_TYPE_DNSKEY",
	DNS_TYPE_DHCID:      "DNS_TYPE_DHCID",
	DNS_TYPE_NSEC3:      "DNS_TYPE_NSEC3",
	DNS_TYPE_NSEC3PARAM: "DNS_TYPE_NSEC3PARAM",
	DNS_TYPE_TLSA:       "DNS_TYPE_TLSA",
	DNS_TYPE_ALL:        "DNS_TYPE_ALL",
	DNS_TYPE_WINS:       "DNS_TYPE_WINS",
	DNS_TYPE_WINSR:      "DNS_TYPE_WINSR",
}

// String returns the [MS-DNSP] constant name of the record type (for example "DNS_TYPE_A"),
// or a hexadecimal fallback of the form "RecordType(0x1234)" for values not defined in the
// specification's table (which still MUST be enumerable per section 2.2.2.1.1).
func (t RecordType) String() string {
	if name, ok := recordTypeNames[t]; ok {
		return name
	}
	const hexDigits = "0123456789abcdef"
	buf := []byte("RecordType(0x0000)")
	buf[13] = hexDigits[(t>>12)&0xF]
	buf[14] = hexDigits[(t>>8)&0xF]
	buf[15] = hexDigits[(t>>4)&0xF]
	buf[16] = hexDigits[t&0xF]
	return string(buf)
}
