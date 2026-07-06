package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// CAUH is a counted array of ULARGE_INTEGER ([MS-MQMQ] 2.2.17): cElems followed by a
// [size_is(cElems)] pointer to that many unsigned 64-bit elements. It is the
// VT_VECTOR|VT_UI8 PROPVARIANT arm.
type CAUH struct {
	CElems ndr.DWORD
	PElems []msdtyp.ULARGE_INTEGER `ndr:"unique,size_is=CElems"`
}
