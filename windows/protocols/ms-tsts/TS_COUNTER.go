package mststs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TS_COUNTER_HEADER identifies a Terminal Services performance counter and whether the
// operation on it succeeded ([MS-TSTS] 2.2.2.6.1, allproc.h _TS_COUNTER_HEADER). The IDL
// declares bResult as `boolean`, the NDR 1-octet boolean ([C706] 14.2.4) — not the
// 4-octet Windows BOOL — so it is modeled as ndr.BOOLEAN (Go bool).
type TS_COUNTER_HEADER struct {
	DwCounterID ndr.DWORD
	BResult     ndr.BOOLEAN
}

// TS_COUNTER carries the value of a single Terminal Services performance counter
// ([MS-TSTS] 2.2.2.6.2, allproc.h _TS_COUNTER).
type TS_COUNTER struct {
	CounterHead TS_COUNTER_HEADER
	DwValue     ndr.DWORD
	StartTime   dtyp.LARGE_INTEGER
}
