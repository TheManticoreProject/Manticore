package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarCreateTrustedDomainEx2Request is the [in] parameter set of
// LsarCreateTrustedDomainEx2: an open policy handle, the extended trusted-domain
// information, the internal (encrypted) authentication information, and the desired access
// mask. Both information parameters are top-level [ref] structs, so they are inlined.
type lsarCreateTrustedDomainEx2Request struct {
	PolicyHandle              structures.LSAPR_HANDLE
	TrustedDomainInformation  structures.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX
	AuthenticationInformation structures.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL
	DesiredAccess             ndr.DWORD
}

func (*lsarCreateTrustedDomainEx2Request) Opnum() uint16 {
	return lsarpc.OpnumLsarCreateTrustedDomainEx2
}

// LsarCreateTrustedDomainEx2 calls LsarCreateTrustedDomainEx2 (opnum 59), creating a
// trusted domain object from the supplied extended trust and internal (encrypted)
// authentication information and returning a handle to it.
func LsarCreateTrustedDomainEx2(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainInformation structures.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX, authenticationInformation structures.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarCreateTrustedDomainEx2Request{
		PolicyHandle:              policyHandle,
		TrustedDomainInformation:  trustedDomainInformation,
		AuthenticationInformation: authenticationInformation,
		DesiredAccess:             ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateTrustedDomainEx2: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateTrustedDomainEx2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
