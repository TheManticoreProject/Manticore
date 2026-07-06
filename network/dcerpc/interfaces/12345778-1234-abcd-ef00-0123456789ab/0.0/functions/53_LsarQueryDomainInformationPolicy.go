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

// lsarQueryDomainInformationPolicyRequest is the [in] parameter set of
// LsarQueryDomainInformationPolicy: an open policy handle and the information class
// selecting which union arm is returned.
type lsarQueryDomainInformationPolicyRequest struct {
	PolicyHandle     mslsad.LSAPR_HANDLE
	InformationClass mslsad.POLICY_DOMAIN_INFORMATION_CLASS `ndr:"enum"`
}

func (*lsarQueryDomainInformationPolicyRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryDomainInformationPolicy
}

// lsarQueryDomainInformationPolicyResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryDomainInformationPolicyResponse struct {
	PolicyDomainInformation *mslsad.LSAPR_POLICY_DOMAIN_INFORMATION `ndr:"unique"`
	Status                  ndr.DWORD                               `ndr:"retval"`
}

// LsarQueryDomainInformationPolicy calls LsarQueryDomainInformationPolicy (opnum 53),
// returning the requested policy domain information class ([MS-LSAD] 3.1.4.4.7).
func LsarQueryDomainInformationPolicy(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, infoClass mslsad.POLICY_DOMAIN_INFORMATION_CLASS) (*mslsad.LSAPR_POLICY_DOMAIN_INFORMATION, error) {
	req := &lsarQueryDomainInformationPolicyRequest{
		PolicyHandle:     policyHandle,
		InformationClass: infoClass,
	}
	var resp lsarQueryDomainInformationPolicyResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryDomainInformationPolicy: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.PolicyDomainInformation, fmt.Errorf("LsarQueryDomainInformationPolicy failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.PolicyDomainInformation, nil
}
