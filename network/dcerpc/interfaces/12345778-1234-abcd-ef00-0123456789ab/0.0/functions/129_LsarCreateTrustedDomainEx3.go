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

// lsarCreateTrustedDomainEx3Request carries the [in] parameters of LsarCreateTrustedDomainEx3.
type lsarCreateTrustedDomainEx3Request struct {
	PolicyHandle              mslsad.LSAPR_HANDLE
	TrustedDomainInformation  mslsad.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX
	AuthenticationInformation mslsad.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES
	DesiredAccess             ndr.DWORD
}

func (*lsarCreateTrustedDomainEx3Request) Opnum() uint16 {
	return lsarpc.OpnumLsarCreateTrustedDomainEx3
}

// lsarCreateTrustedDomainEx3Response carries the [out] parameters and return value of LsarCreateTrustedDomainEx3.
type lsarCreateTrustedDomainEx3Response struct {
	TrustedDomainHandle mslsad.LSAPR_HANDLE
	Status              ndr.DWORD `ndr:"retval"`
}

// LsarCreateTrustedDomainEx3 calls LsarCreateTrustedDomainEx3 (opnum 129) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarCreateTrustedDomainEx3(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainInformation mslsad.LSAPR_TRUSTED_DOMAIN_INFORMATION_EX, authenticationInformation mslsad.LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES, desiredAccess ndr.DWORD) (TrustedDomainHandle mslsad.LSAPR_HANDLE, err error) {
	req := &lsarCreateTrustedDomainEx3Request{
		PolicyHandle:              policyHandle,
		TrustedDomainInformation:  trustedDomainInformation,
		AuthenticationInformation: authenticationInformation,
		DesiredAccess:             desiredAccess,
	}
	var resp lsarCreateTrustedDomainEx3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarCreateTrustedDomainEx3: %w", err)
		return
	}
	TrustedDomainHandle = resp.TrustedDomainHandle
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarCreateTrustedDomainEx3 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
