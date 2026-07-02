package mseven6

// EvtRpcAssertConfigFlags is a [v1_enum] NDR enum ([MS-EVEN6] 2.2.5), transmitted as a
// 32-bit value ([C706] 14.3.6: v1_enum enumerations are sent as unsigned long, not the
// default 16-bit form).
type EvtRpcAssertConfigFlags uint32

const (
	EvtRpcChannelPath   EvtRpcAssertConfigFlags = 0
	EvtRpcPublisherName EvtRpcAssertConfigFlags = 1
)
