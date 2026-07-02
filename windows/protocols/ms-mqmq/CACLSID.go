package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CACLSID is a counted array of CLSID/GUID ([MS-MQMQ] 2.2.19): cElems followed by a
// [size_is(cElems)] pointer to that many GUIDs. It is the VT_VECTOR|VT_CLSID PROPVARIANT
// arm. The elements are dtyp.GUID so each marshals as its 16 octets ([MS-DTYP] 2.3.4.2).
type CACLSID struct {
	CElems ndr.DWORD
	PElems []dtyp.GUID `ndr:"unique,size_is=CElems"`
}
