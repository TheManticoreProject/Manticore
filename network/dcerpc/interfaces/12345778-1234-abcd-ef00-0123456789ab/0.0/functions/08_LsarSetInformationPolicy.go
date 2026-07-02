package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetInformationPolicyRequest is the [in] parameter set of LsarSetInformationPolicy: an
// open policy handle, the information class, and the inline [ref] union value carrying the
// new policy information.
type lsarSetInformationPolicyRequest struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	InformationClass  mslsad.POLICY_INFORMATION_CLASS `ndr:"enum"`
	PolicyInformation mslsad.LSAPR_POLICY_INFORMATION
}

func (*lsarSetInformationPolicyRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetInformationPolicy
}

// LsarSetInformationPolicy calls LsarSetInformationPolicy (opnum 8), setting the policy
// information for the given class ([MS-LSAD] 3.1.4.4.6). The union discriminant is set to
// the information class before marshalling.
func LsarSetInformationPolicy(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, infoClass mslsad.POLICY_INFORMATION_CLASS, info mslsad.LSAPR_POLICY_INFORMATION) error {
	info.Class = infoClass
	req := &lsarSetInformationPolicyRequest{
		PolicyHandle:      policyHandle,
		InformationClass:  infoClass,
		PolicyInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetInformationPolicy: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetInformationPolicy failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
