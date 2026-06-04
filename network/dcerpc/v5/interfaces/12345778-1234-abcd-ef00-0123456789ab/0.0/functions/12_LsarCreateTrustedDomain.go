package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarCreateTrustedDomainRequest is the [in] parameter set of LsarCreateTrustedDomain:
// an open policy handle, the trusted-domain information (a top-level [ref] struct, so it
// is inlined), and the desired access mask for the returned handle.
type lsarCreateTrustedDomainRequest struct {
	PolicyHandle             structures.LSAPR_HANDLE
	TrustedDomainInformation structures.LSAPR_TRUST_INFORMATION
	DesiredAccess            ndr.DWORD
}

func (*lsarCreateTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarCreateTrustedDomain
}

// LsarCreateTrustedDomain calls LsarCreateTrustedDomain (opnum 12), creating a trusted
// domain object from the supplied trust information and returning a handle to it.
func LsarCreateTrustedDomain(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainInformation structures.LSAPR_TRUST_INFORMATION, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarCreateTrustedDomainRequest{
		PolicyHandle:             policyHandle,
		TrustedDomainInformation: trustedDomainInformation,
		DesiredAccess:            ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
