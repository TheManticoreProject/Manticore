package msfrs2

// UPDATE_STATUS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FRS2]).
type UPDATE_STATUS uint16

const (
	UPDATE_STATUS_DONE UPDATE_STATUS = 2
	UPDATE_STATUS_MORE UPDATE_STATUS = 3
)
