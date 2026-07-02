package msdssp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DSROLER_PRIMARY_DOMAIN_INFO_BASIC models DSROLER_PRIMARY_DOMAIN_INFO_BASIC ([MS-DSSP] 2.2.3).
// It carries the domain-related state of the machine when the info level is
// DsRolePrimaryDomainInfoBasic. The three domain-name pointers are [unique, string]
// wide-character strings and may be null. DomainGuid is the [MS-DTYP] GUID (16 octets on
// the wire); dtyp.GUID mirrors that layout exactly (windows/guid.GUID would marshal as 24).
type DSROLER_PRIMARY_DOMAIN_INFO_BASIC struct {
	MachineRole      DSROLE_MACHINE_ROLE
	Flags            ndr.DWORD
	DomainNameFlat   *ndr.WSTR `ndr:"unique"`
	DomainNameDns    *ndr.WSTR `ndr:"unique"`
	DomainForestName *ndr.WSTR `ndr:"unique"`
	DomainGuid       dtyp.GUID
}
