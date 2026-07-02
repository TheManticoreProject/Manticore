package msfasp

// FW_DIRECTION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_DIRECTION uint16

const (
	FW_DIR_INVALID FW_DIRECTION = 0
	FW_DIR_IN      FW_DIRECTION = 1
	FW_DIR_OUT     FW_DIRECTION = 2
	FW_DIR_MAX     FW_DIRECTION = 3
)
