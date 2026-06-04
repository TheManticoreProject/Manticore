package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_503_I contains information about a shared resource, including the server
// name and a security descriptor ([MS-SRVS] 2.2.4.27). The string fields are embedded
// [unique] [string] WCHAR*. shi503_security_descriptor is a [unique]
// [size_is(shi503_reserved)] PUCHAR security-descriptor blob.
type SHARE_INFO_503_I struct {
	Shi503Netname            ndr.WSTR `ndr:"unique"`
	Shi503Type               ndr.DWORD
	Shi503Remark             ndr.WSTR `ndr:"unique"`
	Shi503Permissions        ndr.DWORD
	Shi503MaxUses            ndr.DWORD
	Shi503CurrentUses        ndr.DWORD
	Shi503Path               ndr.WSTR `ndr:"unique"`
	Shi503Passwd             ndr.WSTR `ndr:"unique"`
	Shi503Servername         ndr.WSTR `ndr:"unique"`
	Shi503Reserved           ndr.DWORD
	Shi503SecurityDescriptor []byte `ndr:"unique,size_is=Shi503Reserved"`
}
