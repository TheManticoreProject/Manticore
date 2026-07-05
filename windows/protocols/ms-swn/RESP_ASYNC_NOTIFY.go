package msswn

// RESP_ASYNC_NOTIFY is the notification response returned by WitnessrAsyncNotify
// ([MS-SWN] 2.2.2.4). MessageType selects the shape encoded in MessageBuffer
// (RESOURCE_CHANGE_NOTIFICATION, CLIENT_MOVE_NOTIFICATION, SHARE_MOVE_NOTIFICATION,
// IP_CHANGE_NOTIFICATION); NumberOfMessages is the number of encoded messages; Length is
// the byte length of MessageBuffer.
//
// MessageBuffer is a [size_is(Length)] [unique] PBYTE: a [unique] pointer to a conformant
// byte array sized by Length. Its contents are an opaque, message-type-specific blob that
// callers parse per [MS-SWN] 2.2.2 (this codec transports it as raw bytes).
type RESP_ASYNC_NOTIFY struct {
	MessageType      uint32
	Length           uint32
	NumberOfMessages uint32
	MessageBuffer    []uint8 `ndr:"unique,size_is=Length"`
}
