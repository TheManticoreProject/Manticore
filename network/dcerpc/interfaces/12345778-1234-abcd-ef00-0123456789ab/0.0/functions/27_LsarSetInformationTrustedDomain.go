package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarSetInformationTrustedDomainRequest is the [in] parameter set of
// LsarSetInformationTrustedDomain: an open trusted-domain handle, the information class, and
// the inline [ref] union value carrying the new trusted-domain information.
type lsarSetInformationTrustedDomainRequest struct {
	TrustedDomainHandle      structures.LSAPR_HANDLE
	InformationClass         structures.TRUSTED_INFORMATION_CLASS `ndr:"enum"`
	TrustedDomainInformation structures.LSAPR_TRUSTED_DOMAIN_INFO
}

func (*lsarSetInformationTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetInformationTrustedDomain
}

// LsarSetInformationTrustedDomain calls LsarSetInformationTrustedDomain (opnum 27), setting
// the trusted-domain information for the given class ([MS-LSAD] 3.1.4.7.5). The union
// discriminant is set to the information class before marshalling.
func LsarSetInformationTrustedDomain(rpc ndr.Invoker, trustedDomainHandle structures.LSAPR_HANDLE, infoClass structures.TRUSTED_INFORMATION_CLASS, info structures.LSAPR_TRUSTED_DOMAIN_INFO) error {
	info.Class = infoClass
	req := &lsarSetInformationTrustedDomainRequest{
		TrustedDomainHandle:      trustedDomainHandle,
		InformationClass:         infoClass,
		TrustedDomainInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetInformationTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetInformationTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
