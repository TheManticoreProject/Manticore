package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_SR_SECURITY_DESCRIPTOR is a self-relative security descriptor carried as an
// opaque byte blob ([MS-LSAD] 2.2.3.1). SecurityDescriptor is a [unique] pointer to a
// conformant byte array sized by Length.
type LSAPR_SR_SECURITY_DESCRIPTOR struct {
	Length             ndr.DWORD
	SecurityDescriptor []byte `ndr:"unique,size_is=Length"`
}
