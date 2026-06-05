package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarSetDomainInformationPolicyRequest is the [in] parameter set of
// LsarSetDomainInformationPolicy: an open policy handle, the information class, and the
// [in, unique] union pointer carrying the new policy domain information (NULL deletes it).
type lsarSetDomainInformationPolicyRequest struct {
	PolicyHandle            structures.LSAPR_HANDLE
	InformationClass        structures.POLICY_DOMAIN_INFORMATION_CLASS
	PolicyDomainInformation *structures.LSAPR_POLICY_DOMAIN_INFORMATION `ndr:"unique"`
}

func (*lsarSetDomainInformationPolicyRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetDomainInformationPolicy
}

// LsarSetDomainInformationPolicy calls LsarSetDomainInformationPolicy (opnum 54), setting the
// policy domain information for the given class ([MS-LSAD] 3.1.4.4.8). When info is non-nil
// its union discriminant is set to the information class before marshalling.
func LsarSetDomainInformationPolicy(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, infoClass structures.POLICY_DOMAIN_INFORMATION_CLASS, info *structures.LSAPR_POLICY_DOMAIN_INFORMATION) error {
	if info != nil {
		info.Class = infoClass
	}
	req := &lsarSetDomainInformationPolicyRequest{
		PolicyHandle:            policyHandle,
		InformationClass:        infoClass,
		PolicyDomainInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetDomainInformationPolicy: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetDomainInformationPolicy failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
