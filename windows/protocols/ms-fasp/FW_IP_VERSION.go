package msfasp

// FW_IP_VERSION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_IP_VERSION uint16

const (
	FW_IP_VERSION_INVALID FW_IP_VERSION = 0
	FW_IP_VERSION_V4      FW_IP_VERSION = 1
	FW_IP_VERSION_V6      FW_IP_VERSION = 2
	FW_IP_VERSION_MAX     FW_IP_VERSION = 3
)
