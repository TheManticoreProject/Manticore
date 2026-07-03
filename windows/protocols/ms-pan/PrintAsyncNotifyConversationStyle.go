package mspan

// PrintAsyncNotifyConversationStyle is a [v1_enum] NDR enum ([MS-PAN] 2.2.2), so it is
// transmitted as a 32-bit value ([C706] 14.3.6, MIDL v1_enum), not the 16-bit default.
type PrintAsyncNotifyConversationStyle uint32

// The constants keep their IDL names ([MS-PAN] 2.2.2) but are exported so callers of the
// IRPCAsyncNotify stubs can name them across the package boundary.
const (
	KBiDirectional  PrintAsyncNotifyConversationStyle = 0
	KUniDirectional PrintAsyncNotifyConversationStyle = 1
)
