package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetTrustedDomainInfoRequest is the [in] parameter set of LsarSetTrustedDomainInfo: an
// open policy handle, the [unique] SID identifying the trusted domain, the information class,
// and the inline [ref] union value carrying the new trusted-domain information.
type lsarSetTrustedDomainInfoRequest struct {
	PolicyHandle             mslsad.LSAPR_HANDLE
	TrustedDomainSid         *msdtyp.RPC_SID                  `ndr:"unique"`
	InformationClass         mslsad.TRUSTED_INFORMATION_CLASS `ndr:"enum"`
	TrustedDomainInformation mslsad.LSAPR_TRUSTED_DOMAIN_INFO
}

func (*lsarSetTrustedDomainInfoRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetTrustedDomainInfo
}

// LsarSetTrustedDomainInfo calls LsarSetTrustedDomainInfo (opnum 40), setting the
// trusted-domain information for the domain identified by SID ([MS-LSAD] 3.1.4.7.3). The
// union discriminant is set to the information class before marshalling.
func LsarSetTrustedDomainInfo(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainSid *msdtyp.RPC_SID, infoClass mslsad.TRUSTED_INFORMATION_CLASS, info mslsad.LSAPR_TRUSTED_DOMAIN_INFO) error {
	info.Class = infoClass
	req := &lsarSetTrustedDomainInfoRequest{
		PolicyHandle:             policyHandle,
		TrustedDomainSid:         trustedDomainSid,
		InformationClass:         infoClass,
		TrustedDomainInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetTrustedDomainInfo: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetTrustedDomainInfo failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
