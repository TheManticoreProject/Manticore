package msdnsp

// DNS_RRL_MODE_ENUM is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RRL_MODE_ENUM uint16

const (
	DnsRRLLogOnly  DNS_RRL_MODE_ENUM = 0
	DnsRRLEnabled  DNS_RRL_MODE_ENUM = 1
	DnsRRLDisabled DNS_RRL_MODE_ENUM = 2
)
