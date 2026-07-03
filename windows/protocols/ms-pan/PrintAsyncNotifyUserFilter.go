package mspan

// PrintAsyncNotifyUserFilter is a [v1_enum] NDR enum ([MS-PAN] 2.2.3), so it is
// transmitted as a 32-bit value ([C706] 14.3.6, MIDL v1_enum), not the 16-bit default.
type PrintAsyncNotifyUserFilter uint32

// The constants keep their IDL names ([MS-PAN] 2.2.3) but are exported so callers of the
// IRPCAsyncNotify stubs can name them across the package boundary.
const (
	KPerUser  PrintAsyncNotifyUserFilter = 0
	KAllUsers PrintAsyncNotifyUserFilter = 1
)
