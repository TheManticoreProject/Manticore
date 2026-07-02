package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CAUI is a counted array of unsigned short ([MS-MQMQ] 2.2.14): cElems followed by a
// [size_is(cElems)] pointer to that many 16-bit elements. It is the VT_VECTOR|VT_UI2 arm.
type CAUI struct {
	CElems ndr.DWORD
	PElems []uint16 `ndr:"unique,size_is=CElems"`
}
