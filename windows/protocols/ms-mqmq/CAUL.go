package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CAUL is a counted array of unsigned long ([MS-MQMQ] 2.2.16): cElems followed by a
// [size_is(cElems)] pointer to that many unsigned 32-bit elements. It is the
// VT_VECTOR|VT_UI4 PROPVARIANT arm.
type CAUL struct {
	CElems ndr.DWORD
	PElems []ndr.DWORD `ndr:"unique,size_is=CElems"`
}
