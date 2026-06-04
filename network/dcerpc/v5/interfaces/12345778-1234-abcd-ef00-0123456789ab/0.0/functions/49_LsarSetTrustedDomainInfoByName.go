package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarSetTrustedDomainInfoByNameRequest is the [in] parameter set of
// LsarSetTrustedDomainInfoByName: an open policy handle, the [unique] trusted-domain name,
// the information class, and the inline [ref] union value carrying the new trusted-domain
// information.
type lsarSetTrustedDomainInfoByNameRequest struct {
	PolicyHandle             structures.LSAPR_HANDLE
	TrustedDomainName        *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
	InformationClass         structures.TRUSTED_INFORMATION_CLASS
	TrustedDomainInformation structures.LSAPR_TRUSTED_DOMAIN_INFO
}

func (*lsarSetTrustedDomainInfoByNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetTrustedDomainInfoByName
}

// LsarSetTrustedDomainInfoByName calls LsarSetTrustedDomainInfoByName (opnum 49), setting the
// trusted-domain information for the domain identified by name ([MS-LSAD] 3.1.4.7.7). The
// union discriminant is set to the information class before marshalling.
func LsarSetTrustedDomainInfoByName(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainName string, infoClass structures.TRUSTED_INFORMATION_CLASS, info structures.LSAPR_TRUSTED_DOMAIN_INFO) error {
	info.Class = infoClass
	name := dtyp.NewUnicodeString(trustedDomainName)
	req := &lsarSetTrustedDomainInfoByNameRequest{
		PolicyHandle:             policyHandle,
		TrustedDomainName:        &name,
		InformationClass:         infoClass,
		TrustedDomainInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetTrustedDomainInfoByName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetTrustedDomainInfoByName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
