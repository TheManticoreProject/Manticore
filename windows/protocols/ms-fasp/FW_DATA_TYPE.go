package msfasp

// FW_DATA_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_DATA_TYPE uint16

const (
	FW_DATA_TYPE_EMPTY          FW_DATA_TYPE = 0
	FW_DATA_TYPE_UINT8          FW_DATA_TYPE = 1
	FW_DATA_TYPE_UINT16         FW_DATA_TYPE = 2
	FW_DATA_TYPE_UINT32         FW_DATA_TYPE = 3
	FW_DATA_TYPE_UINT64         FW_DATA_TYPE = 4
	FW_DATA_TYPE_UNICODE_STRING FW_DATA_TYPE = 5
)
