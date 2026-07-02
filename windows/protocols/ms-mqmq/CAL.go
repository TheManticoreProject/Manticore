package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CAL is a counted array of long ([MS-MQMQ] 2.2.15): cElems followed by a
// [size_is(cElems)] pointer to that many signed 32-bit elements. It is the
// VT_VECTOR|VT_I4 PROPVARIANT arm.
type CAL struct {
	CElems ndr.DWORD
	PElems []int32 `ndr:"unique,size_is=CElems"`
}
