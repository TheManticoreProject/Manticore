package msmqmp

// TRANSFER_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-MQMP]).
type TRANSFER_TYPE uint16

const (
	CACTB_SEND         TRANSFER_TYPE = 0
	CACTB_RECEIVE      TRANSFER_TYPE = 1
	CACTB_CREATECURSOR TRANSFER_TYPE = 2
)
