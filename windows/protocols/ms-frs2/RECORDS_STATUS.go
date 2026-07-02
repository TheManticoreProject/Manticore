package msfrs2

// RECORDS_STATUS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FRS2]).
type RECORDS_STATUS uint16

const (
	RECORDS_STATUS_DONE RECORDS_STATUS = 0
	RECORDS_STATUS_MORE RECORDS_STATUS = 1
)
