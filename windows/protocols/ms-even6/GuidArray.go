package mseven6

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// GuidArray ([MS-EVEN6] 2.2.9) is a counted array of GUIDs. Ptr is a [unique] pointer to a
// conformant array whose length is Count. The element is dtyp.GUID (16 octets: Data1/2/3 +
// Data4[8]); windows/guid.GUID must not be used on the wire — its uint64 tail marshals to
// 24 bytes under NDR.
type GuidArray struct {
	Count ndr.DWORD
	Ptr   []dtyp.GUID `ndr:"unique,size_is=Count"`
}
