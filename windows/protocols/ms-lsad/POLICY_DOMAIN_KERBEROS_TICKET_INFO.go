package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// POLICY_DOMAIN_KERBEROS_TICKET_INFO communicates policy information about the Kerberos
// security provider ([MS-LSAD] 2.2.4.20).
type POLICY_DOMAIN_KERBEROS_TICKET_INFO struct {
	AuthenticationOptions ndr.DWORD
	MaxServiceTicketAge   msdtyp.LARGE_INTEGER
	MaxTicketAge          msdtyp.LARGE_INTEGER
	MaxRenewAge           msdtyp.LARGE_INTEGER
	MaxClockSkew          msdtyp.LARGE_INTEGER
	Reserved              msdtyp.LARGE_INTEGER
}
