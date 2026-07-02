package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarQueryInfoTrustedDomainRequest is the [in] parameter set of LsarQueryInfoTrustedDomain:
// an open trusted-domain handle and the information class selecting which union arm is
// returned.
type lsarQueryInfoTrustedDomainRequest struct {
	TrustedDomainHandle mslsad.LSAPR_HANDLE
	InformationClass    mslsad.TRUSTED_INFORMATION_CLASS `ndr:"enum"`
}

func (*lsarQueryInfoTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryInfoTrustedDomain
}

// lsarQueryInfoTrustedDomainResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryInfoTrustedDomainResponse struct {
	TrustedDomainInformation *mslsad.LSAPR_TRUSTED_DOMAIN_INFO `ndr:"unique"`
	Status                   ndr.DWORD                         `ndr:"retval"`
}

// LsarQueryInfoTrustedDomain calls LsarQueryInfoTrustedDomain (opnum 26), returning the
// requested trusted-domain information class ([MS-LSAD] 3.1.4.7.4).
func LsarQueryInfoTrustedDomain(rpc ndr.Invoker, trustedDomainHandle mslsad.LSAPR_HANDLE, infoClass mslsad.TRUSTED_INFORMATION_CLASS) (*mslsad.LSAPR_TRUSTED_DOMAIN_INFO, error) {
	req := &lsarQueryInfoTrustedDomainRequest{
		TrustedDomainHandle: trustedDomainHandle,
		InformationClass:    infoClass,
	}
	var resp lsarQueryInfoTrustedDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryInfoTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.TrustedDomainInformation, fmt.Errorf("LsarQueryInfoTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.TrustedDomainInformation, nil
}
