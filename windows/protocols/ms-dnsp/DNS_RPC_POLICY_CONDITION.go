package msdnsp

// DNS_RPC_POLICY_CONDITION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_POLICY_CONDITION uint16

const (
	DNS_AND DNS_RPC_POLICY_CONDITION = 0
	DNS_OR  DNS_RPC_POLICY_CONDITION = 1
)
