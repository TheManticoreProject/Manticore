package msfasp

// FW_OBJECT_CTRL_FLAG is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_OBJECT_CTRL_FLAG uint16

const (
	FW_OBJECT_CTRL_FLAG_INCLUDE_METADATA FW_OBJECT_CTRL_FLAG = 0x0001
)
