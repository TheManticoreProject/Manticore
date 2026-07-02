package msdnsp

// DNS_RPC_CRITERIA_COMPARATOR is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_CRITERIA_COMPARATOR uint16

const (
	Equals    DNS_RPC_CRITERIA_COMPARATOR = 1
	NotEquals DNS_RPC_CRITERIA_COMPARATOR = 2
)
