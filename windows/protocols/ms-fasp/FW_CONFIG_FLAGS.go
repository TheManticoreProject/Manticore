package msfasp

// FW_CONFIG_FLAGS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_CONFIG_FLAGS uint16

const (
	FW_CONFIG_FLAG_RETURN_DEFAULT_IF_NOT_FOUND FW_CONFIG_FLAGS = 0x0001
)
