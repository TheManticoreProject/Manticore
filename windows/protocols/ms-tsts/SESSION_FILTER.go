package mststs

// SESSION_FILTER selects which sessions RpcGetSessionIds enumerates ([MS-TSTS] 2.2.1.2,
// tsdef.h _SESSION_FILTER). NDR transmits an enum as a 16-bit value ([C706] 14.3.6), so
// it is defined as a uint16.
type SESSION_FILTER uint16

const (
	// SF_SERVICES_SESSION_POPUP requests the sessions that can display a services message box.
	SF_SERVICES_SESSION_POPUP SESSION_FILTER = 0
)
