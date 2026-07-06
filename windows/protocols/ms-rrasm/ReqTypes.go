package msrrasm

// ReqTypes is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type ReqTypes uint16

const (
	REQTYPE_PORTENUM             ReqTypes = 21
	REQTYPE_GETINFO              ReqTypes = 22
	REQTYPE_GETDEVCONFIG         ReqTypes = 73
	REQTYPE_SETDEVICECONFIGINFO  ReqTypes = 94
	REQTYPE_GETDEVICECONFIGINFO  ReqTypes = 95
	REQTYPE_GETCALLEDID          ReqTypes = 105
	REQTYPE_SETCALLEDID          ReqTypes = 106
	REQTYPE_GETNDISWANDRIVERCAPS ReqTypes = 111
)
