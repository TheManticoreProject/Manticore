package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CAPROPVARIANT is a counted array of PROPVARIANT ([MS-MQMQ] 2.2.21): cElems followed by
// a [size_is(cElems)] pointer to that many PROPVARIANT elements. It is the
// VT_VECTOR|VT_VARIANT arm of a PROPVARIANT, so the type is (indirectly) recursive.
type CAPROPVARIANT struct {
	CElems ndr.DWORD
	PElems []PROPVARIANT `ndr:"unique,size_is=CElems"`
}
