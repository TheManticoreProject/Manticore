package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ADT_SECURITY_DESCRIPTOR holds a marshalled self-relative security descriptor
// ([MS-SRVS] 2.2.4.83). Buffer is a [unique] pointer to a conformant byte array
// sized by Length.
type ADT_SECURITY_DESCRIPTOR struct {
	Length ndr.DWORD
	Buffer []byte `ndr:"unique,size_is=Length"`
}
