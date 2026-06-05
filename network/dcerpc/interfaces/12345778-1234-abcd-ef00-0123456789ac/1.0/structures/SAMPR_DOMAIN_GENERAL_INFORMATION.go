package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_GENERAL_INFORMATION contains general per-domain attributes
// ([MS-SAMR] 2.2.4.9). DomainServerState and DomainServerRole are transmitted as
// unsigned long per the IDL (not as the enum types), and ForceLogoff and
// DomainModifiedCount are OLD_LARGE_INTEGER values defined by the base family.
type SAMPR_DOMAIN_GENERAL_INFORMATION struct {
	ForceLogoff              OLD_LARGE_INTEGER
	OemInformation           dtyp.RPC_UNICODE_STRING
	DomainName               dtyp.RPC_UNICODE_STRING
	ReplicaSourceNodeName    dtyp.RPC_UNICODE_STRING
	DomainModifiedCount      OLD_LARGE_INTEGER
	DomainServerState        ndr.DWORD
	DomainServerRole         ndr.DWORD
	UasCompatibilityRequired uint8
	UserCount                ndr.DWORD
	GroupCount               ndr.DWORD
	AliasCount               ndr.DWORD
}
