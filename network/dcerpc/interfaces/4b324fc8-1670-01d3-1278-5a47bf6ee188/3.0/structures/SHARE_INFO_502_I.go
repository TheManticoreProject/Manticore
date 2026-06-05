package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_502_I contains information about a shared resource, including its
// security descriptor ([MS-SRVS] 2.2.4.26). The string fields are embedded [unique]
// [string] WCHAR*. shi502_security_descriptor is a [unique] [size_is(shi502_reserved)]
// unsigned char* security-descriptor blob.
type SHARE_INFO_502_I struct {
	Shi502Netname            ndr.WSTR `ndr:"unique"`
	Shi502Type               ndr.DWORD
	Shi502Remark             ndr.WSTR `ndr:"unique"`
	Shi502Permissions        ndr.DWORD
	Shi502MaxUses            ndr.DWORD
	Shi502CurrentUses        ndr.DWORD
	Shi502Path               ndr.WSTR `ndr:"unique"`
	Shi502Passwd             ndr.WSTR `ndr:"unique"`
	Shi502Reserved           ndr.DWORD
	Shi502SecurityDescriptor []byte `ndr:"unique,size_is=Shi502Reserved"`
}
