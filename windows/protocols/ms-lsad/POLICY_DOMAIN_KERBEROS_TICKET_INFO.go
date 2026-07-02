package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// POLICY_DOMAIN_KERBEROS_TICKET_INFO communicates policy information about the Kerberos
// security provider ([MS-LSAD] 2.2.4.20).
type POLICY_DOMAIN_KERBEROS_TICKET_INFO struct {
	AuthenticationOptions ndr.DWORD
	MaxServiceTicketAge   dtyp.LARGE_INTEGER
	MaxTicketAge          dtyp.LARGE_INTEGER
	MaxRenewAge           dtyp.LARGE_INTEGER
	MaxClockSkew          dtyp.LARGE_INTEGER
	Reserved              dtyp.LARGE_INTEGER
}
