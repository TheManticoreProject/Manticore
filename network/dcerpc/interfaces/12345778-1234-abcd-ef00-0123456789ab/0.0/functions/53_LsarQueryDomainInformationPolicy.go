package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarQueryDomainInformationPolicyRequest is the [in] parameter set of
// LsarQueryDomainInformationPolicy: an open policy handle and the information class
// selecting which union arm is returned.
type lsarQueryDomainInformationPolicyRequest struct {
	PolicyHandle     structures.LSAPR_HANDLE
	InformationClass structures.POLICY_DOMAIN_INFORMATION_CLASS `ndr:"enum"`
}

func (*lsarQueryDomainInformationPolicyRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryDomainInformationPolicy
}

// lsarQueryDomainInformationPolicyResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryDomainInformationPolicyResponse struct {
	PolicyDomainInformation *structures.LSAPR_POLICY_DOMAIN_INFORMATION `ndr:"unique"`
	Status                  ndr.DWORD                                   `ndr:"retval"`
}

// LsarQueryDomainInformationPolicy calls LsarQueryDomainInformationPolicy (opnum 53),
// returning the requested policy domain information class ([MS-LSAD] 3.1.4.4.7).
func LsarQueryDomainInformationPolicy(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, infoClass structures.POLICY_DOMAIN_INFORMATION_CLASS) (*structures.LSAPR_POLICY_DOMAIN_INFORMATION, error) {
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
