package mststs

// SESSION_CHANGE describes a single session-state change delivered by
// RpcWaitAsyncNotification ([MS-TSTS] 2.2.2.4, tsdef.h _SESSION_CHANGE). Both fields are
// 4-octet scalars, so the structure has a fixed 8-byte NDR layout.
type SESSION_CHANGE struct {
	SessionId      int32
	NotificationId TNotificationId
}
