package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetInformationPolicy2Request is the [in] parameter set of LsarSetInformationPolicy2: an
// open policy handle, the information class, and the inline [ref] union value carrying the
// new policy information.
type lsarSetInformationPolicy2Request struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	InformationClass  mslsad.POLICY_INFORMATION_CLASS `ndr:"enum"`
	PolicyInformation mslsad.LSAPR_POLICY_INFORMATION
}

func (*lsarSetInformationPolicy2Request) Opnum() uint16 {
	return lsarpc.OpnumLsarSetInformationPolicy2
}

// LsarSetInformationPolicy2 calls LsarSetInformationPolicy2 (opnum 47), setting the policy
// information for the given class ([MS-LSAD] 3.1.4.4.5). The union discriminant is set to
// the information class before marshalling.
func LsarSetInformationPolicy2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, infoClass mslsad.POLICY_INFORMATION_CLASS, info mslsad.LSAPR_POLICY_INFORMATION) error {
	info.Class = infoClass
	req := &lsarSetInformationPolicy2Request{
		PolicyHandle:      policyHandle,
		InformationClass:  infoClass,
		PolicyInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetInformationPolicy2: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetInformationPolicy2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
