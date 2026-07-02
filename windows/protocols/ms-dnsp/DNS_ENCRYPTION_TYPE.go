package msdnsp

// DNS_ENCRYPTION_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_ENCRYPTION_TYPE uint16

const (
	DnsEncryptionNone DNS_ENCRYPTION_TYPE = 0
	DnsEncryptionDoH  DNS_ENCRYPTION_TYPE = 1
)
