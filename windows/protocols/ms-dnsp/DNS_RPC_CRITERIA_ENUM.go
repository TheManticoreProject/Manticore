package msdnsp

// DNS_RPC_CRITERIA_ENUM is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type DNS_RPC_CRITERIA_ENUM uint16

const (
	DnsPolicyCriteriaSubnet            DNS_RPC_CRITERIA_ENUM = 0
	DnsPolicyCriteriaTransportProtocol DNS_RPC_CRITERIA_ENUM = 1
	DnsPolicyCriteriaNetworkProtocol   DNS_RPC_CRITERIA_ENUM = 2
	DnsPolicyCriteriaInterface         DNS_RPC_CRITERIA_ENUM = 3
	DnsPolicyCriteriaFqdn              DNS_RPC_CRITERIA_ENUM = 4
	DnsPolicyCriteriaQtype             DNS_RPC_CRITERIA_ENUM = 5
	DnsPolicyCriteriaTime              DNS_RPC_CRITERIA_ENUM = 6
	DnsPolicyCriteriaMax               DNS_RPC_CRITERIA_ENUM = 7
)
