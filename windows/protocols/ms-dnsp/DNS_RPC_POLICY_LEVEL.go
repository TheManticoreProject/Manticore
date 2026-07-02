package msdnsp

// DNS_RPC_POLICY_LEVEL is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_POLICY_LEVEL uint16

const (
	DnsPolicyServerLevel DNS_RPC_POLICY_LEVEL = 0
	DnsPolicyZoneLevel   DNS_RPC_POLICY_LEVEL = 1
	DnsPolicyLevelMax    DNS_RPC_POLICY_LEVEL = 2
)
