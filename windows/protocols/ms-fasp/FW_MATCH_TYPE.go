package msfasp

// FW_MATCH_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_MATCH_TYPE uint16

const (
	FW_MATCH_TYPE_TRAFFIC_MATCH FW_MATCH_TYPE = 0
	FW_MATCH_TYPE_EQUAL         FW_MATCH_TYPE = 1
	FW_MATCH_TYPE_MAX           FW_MATCH_TYPE = 2
)
