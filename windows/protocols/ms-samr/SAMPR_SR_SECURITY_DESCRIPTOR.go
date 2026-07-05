package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_SR_SECURITY_DESCRIPTOR holds a self-relative security descriptor
// ([MS-SAMR] 2.2.3.12). SecurityDescriptor is a [unique] pointer to a conformant
// array of bytes sized by Length.
type SAMPR_SR_SECURITY_DESCRIPTOR struct {
	Length             ndr.DWORD
	SecurityDescriptor []byte `ndr:"unique,size_is=Length"`
}
