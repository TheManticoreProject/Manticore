package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarQueryInformationPolicyRequest is the [in] parameter set of LsarQueryInformationPolicy:
// an open policy handle and the information class selecting which union arm is returned.
type lsarQueryInformationPolicyRequest struct {
	PolicyHandle     structures.LSAPR_HANDLE
	InformationClass structures.POLICY_INFORMATION_CLASS
}

func (*lsarQueryInformationPolicyRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryInformationPolicy
}

// lsarQueryInformationPolicyResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryInformationPolicyResponse struct {
	PolicyInformation *structures.LSAPR_POLICY_INFORMATION `ndr:"unique"`
	Status            ndr.DWORD
}

// LsarQueryInformationPolicy calls LsarQueryInformationPolicy (opnum 7), returning the
// requested policy information class ([MS-LSAD] 3.1.4.4.4).
func LsarQueryInformationPolicy(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, infoClass structures.POLICY_INFORMATION_CLASS) (*structures.LSAPR_POLICY_INFORMATION, error) {
	req := &lsarQueryInformationPolicyRequest{
		PolicyHandle:     policyHandle,
		InformationClass: infoClass,
	}
	var resp lsarQueryInformationPolicyResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryInformationPolicy: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.PolicyInformation, fmt.Errorf("LsarQueryInformationPolicy failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.PolicyInformation, nil
}
