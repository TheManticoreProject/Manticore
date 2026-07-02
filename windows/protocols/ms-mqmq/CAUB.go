package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CAUB is a counted array of unsigned char ([MS-MQMQ] 2.2.13): cElems followed by a
// [size_is(cElems)] pointer to that many octets. It is the VT_VECTOR|VT_UI1 PROPVARIANT arm.
type CAUB struct {
	CElems ndr.DWORD
	PElems []uint8 `ndr:"unique,size_is=CElems"`
}
