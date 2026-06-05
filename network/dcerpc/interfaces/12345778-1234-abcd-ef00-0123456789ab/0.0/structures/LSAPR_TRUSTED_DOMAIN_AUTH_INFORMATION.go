package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION holds the inbound and outbound authentication
// information for a trust ([MS-LSAD] 2.2.7.11). Each authentication-information member is
// a single [unique] pointer to an LSAPR_AUTH_INFORMATION.
type LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION struct {
	IncomingAuthInfos                         ndr.DWORD
	IncomingAuthenticationInformation         *LSAPR_AUTH_INFORMATION `ndr:"unique"`
	IncomingPreviousAuthenticationInformation *LSAPR_AUTH_INFORMATION `ndr:"unique"`
	OutgoingAuthInfos                         ndr.DWORD
	OutgoingAuthenticationInformation         *LSAPR_AUTH_INFORMATION `ndr:"unique"`
	OutgoingPreviousAuthenticationInformation *LSAPR_AUTH_INFORMATION `ndr:"unique"`
}
