package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrQuerySecurityObjectRequest carries the [in] parameters of
// SamrQuerySecurityObject: the target object handle and the SECURITY_INFORMATION
// selecting which parts of the descriptor to retrieve.
type samrQuerySecurityObjectRequest struct {
	ObjectHandle        mssamr.SAMPR_HANDLE
	SecurityInformation ndr.DWORD
}

func (*samrQuerySecurityObjectRequest) Opnum() uint16 { return samr.OpnumSamrQuerySecurityObject }

// samrQuerySecurityObjectResponse carries the [out] double pointer to the
// self-relative security descriptor and the NTSTATUS.
type samrQuerySecurityObjectResponse struct {
	SecurityDescriptor *mssamr.SAMPR_SR_SECURITY_DESCRIPTOR `ndr:"unique"`
	Status             ndr.DWORD                            `ndr:"retval"`
}

// SamrQuerySecurityObject calls SamrQuerySecurityObject (opnum 3), retrieving the
// security descriptor of the referenced object ([MS-SAMR] 3.1.5.12.1).
func SamrQuerySecurityObject(rpc ndr.Invoker, handle mssamr.SAMPR_HANDLE, securityInformation uint32) (*mssamr.SAMPR_SR_SECURITY_DESCRIPTOR, error) {
	req := &samrQuerySecurityObjectRequest{
		ObjectHandle:        handle,
		SecurityInformation: ndr.DWORD(securityInformation),
	}
	var resp samrQuerySecurityObjectResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQuerySecurityObject: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.SecurityDescriptor, fmt.Errorf("SamrQuerySecurityObject failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.SecurityDescriptor, nil
}
