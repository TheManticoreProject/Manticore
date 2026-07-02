package msdnsp

// DNS_RPC_POLICY_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_POLICY_TYPE uint16

const (
	DnsPolicyQueryProcessing DNS_RPC_POLICY_TYPE = 0
	DnsPolicyZoneTransfer    DNS_RPC_POLICY_TYPE = 1
	DnsPolicyDynamicUpdate   DNS_RPC_POLICY_TYPE = 2
	DnsPolicyRecursion       DNS_RPC_POLICY_TYPE = 3
	DnsPolicyMax             DNS_RPC_POLICY_TYPE = 4
)
