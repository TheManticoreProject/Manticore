package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TNotificationId is a 32-bit bitmask of session-change notification types
// (typedef ULONG TNotificationId, [MS-TSTS] Appendix A.6 tsdef.h / 2.2.1.3). It is
// transmitted as a 4-octet value, so it is defined as ndr.DWORD, not an NDR enum.
type TNotificationId = ndr.DWORD

// Session-change notification masks ([MS-TSTS] 2.2.1.3, tsdef.h WTS_NOTIFY_*).
const (
	WTS_NOTIFY_NONE               TNotificationId = 0x00000000
	WTS_NOTIFY_CREATE             TNotificationId = 0x00000001
	WTS_NOTIFY_CONNECT            TNotificationId = 0x00000002
	WTS_NOTIFY_DISCONNECT         TNotificationId = 0x00000004
	WTS_NOTIFY_LOGON              TNotificationId = 0x00000008
	WTS_NOTIFY_LOGOFF             TNotificationId = 0x00000010
	WTS_NOTIFY_SHADOW_START       TNotificationId = 0x00000020
	WTS_NOTIFY_SHADOW_STOP        TNotificationId = 0x00000040
	WTS_NOTIFY_TERMINATE          TNotificationId = 0x00000080
	WTS_NOTIFY_CONSOLE_CONNECT    TNotificationId = 0x00000100
	WTS_NOTIFY_CONSOLE_DISCONNECT TNotificationId = 0x00000200
	WTS_NOTIFY_LOCK               TNotificationId = 0x00000400
	WTS_NOTIFY_UNLOCK             TNotificationId = 0x00000800
	WTS_NOTIFY_ALL                TNotificationId = 0xffffffff
)
