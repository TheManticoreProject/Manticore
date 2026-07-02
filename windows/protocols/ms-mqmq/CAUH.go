package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CAUH is a counted array of ULARGE_INTEGER ([MS-MQMQ] 2.2.17): cElems followed by a
// [size_is(cElems)] pointer to that many unsigned 64-bit elements. It is the
// VT_VECTOR|VT_UI8 PROPVARIANT arm.
type CAUH struct {
	CElems ndr.DWORD
	PElems []dtyp.ULARGE_INTEGER `ndr:"unique,size_is=CElems"`
}
