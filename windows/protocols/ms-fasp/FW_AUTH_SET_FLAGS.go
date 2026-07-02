package msfasp

// FW_AUTH_SET_FLAGS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_AUTH_SET_FLAGS uint16

const (
	FW_AUTH_SET_FLAGS_NONE FW_AUTH_SET_FLAGS = 0x00
	FW_AUTH_SET_FLAGS_MAX  FW_AUTH_SET_FLAGS = 0x01
)
