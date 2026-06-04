package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrSetSecurityObjectRequest carries the [in] parameters of SamrSetSecurityObject:
// the target object handle, the SECURITY_INFORMATION selecting which parts to set,
// and the [ref] self-relative security descriptor (modeled inline).
type samrSetSecurityObjectRequest struct {
	ObjectHandle        structures.SAMPR_HANDLE
	SecurityInformation ndr.DWORD
	SecurityDescriptor  structures.SAMPR_SR_SECURITY_DESCRIPTOR
}

func (*samrSetSecurityObjectRequest) Opnum() uint16 { return samr.OpnumSamrSetSecurityObject }

// SamrSetSecurityObject calls SamrSetSecurityObject (opnum 2), setting the security
// descriptor of the referenced object ([MS-SAMR] 3.1.5.12.2).
func SamrSetSecurityObject(rpc *client.Client, handle structures.SAMPR_HANDLE, securityInformation uint32, securityDescriptor structures.SAMPR_SR_SECURITY_DESCRIPTOR) error {
	req := &samrSetSecurityObjectRequest{
		ObjectHandle:        handle,
		SecurityInformation: ndr.DWORD(securityInformation),
		SecurityDescriptor:  securityDescriptor,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetSecurityObject: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetSecurityObject failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
