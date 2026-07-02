package msdnsp

// DNS_RPC_POLICY_ACTION_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_POLICY_ACTION_TYPE uint16

const (
	DnsPolicyDeny      DNS_RPC_POLICY_ACTION_TYPE = 0
	DnsPolicyAllow     DNS_RPC_POLICY_ACTION_TYPE = 1
	DnsPolicyIgnore    DNS_RPC_POLICY_ACTION_TYPE = 2
	DnsPolicyActionMax DNS_RPC_POLICY_ACTION_TYPE = 3
)
