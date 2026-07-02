package mseven6

// EvtRpcVariantType is a [v1_enum] NDR enum ([MS-EVEN6] 2.2.4), transmitted as a 32-bit
// value ([C706] 14.3.6: v1_enum enumerations are sent as unsigned long, not the default
// 16-bit form).
type EvtRpcVariantType uint32

const (
	EvtRpcVarTypeNull         EvtRpcVariantType = 0
	EvtRpcVarTypeBoolean      EvtRpcVariantType = 1
	EvtRpcVarTypeUInt32       EvtRpcVariantType = 2
	EvtRpcVarTypeUInt64       EvtRpcVariantType = 3
	EvtRpcVarTypeString       EvtRpcVariantType = 4
	EvtRpcVarTypeGuid         EvtRpcVariantType = 5
	EvtRpcVarTypeBooleanArray EvtRpcVariantType = 6
	EvtRpcVarTypeUInt32Array  EvtRpcVariantType = 7
	EvtRpcVarTypeUInt64Array  EvtRpcVariantType = 8
	EvtRpcVarTypeStringArray  EvtRpcVariantType = 9
	EvtRpcVarTypeGuidArray    EvtRpcVariantType = 10
)
