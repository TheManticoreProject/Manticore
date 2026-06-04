package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1501_I contains a self-relative security descriptor for a shared resource
// ([MS-SRVS] 2.2.4.31). shi1501_security_descriptor is a [unique]
// [size_is(shi1501_reserved)] unsigned char* security-descriptor blob.
type SHARE_INFO_1501_I struct {
	Shi1501Reserved           ndr.DWORD
	Shi1501SecurityDescriptor []byte `ndr:"unique,size_is=Shi1501Reserved"`
}
