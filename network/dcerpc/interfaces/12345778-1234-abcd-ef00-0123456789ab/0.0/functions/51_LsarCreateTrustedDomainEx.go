package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarCreateTrustedDomainExRequest is the [in] parameter set of LsarCreateTrustedDomainEx:
// an open policy handle, the extended trusted-domain information, the authentication
// information, and the desired access mask. Both information parameters are top-level
// [ref] structs, so they are inlined.
type lsarCreateTrustedDomainExRequest struct {
	PolicyHandle              mslsad.LSAPR_HANDLE
	TrustedDomainInformation  mslsad.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX
	AuthenticationInformation mslsad.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION
	DesiredAccess             ndr.DWORD
}

func (*lsarCreateTrustedDomainExRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarCreateTrustedDomainEx
}

// LsarCreateTrustedDomainEx calls LsarCreateTrustedDomainEx (opnum 51), creating a trusted
// domain object from the supplied extended trust and authentication information and
// returning a handle to it.
func LsarCreateTrustedDomainEx(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainInformation mslsad.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX, authenticationInformation mslsad.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarCreateTrustedDomainExRequest{
		PolicyHandle:              policyHandle,
		TrustedDomainInformation:  trustedDomainInformation,
		AuthenticationInformation: authenticationInformation,
		DesiredAccess:             ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateTrustedDomainEx: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateTrustedDomainEx failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
