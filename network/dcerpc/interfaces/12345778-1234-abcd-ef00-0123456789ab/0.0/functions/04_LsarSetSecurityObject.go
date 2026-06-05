package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarSetSecurityObjectRequest is the [in] parameter set of LsarSetSecurityObject: an open
// object handle, the SECURITY_INFORMATION bitmask, and the new self-relative security
// descriptor. SecurityDescriptor is a single top-level [ref] pointer, so it is inlined.
type lsarSetSecurityObjectRequest struct {
	ObjectHandle        structures.LSAPR_HANDLE
	SecurityInformation ndr.DWORD
	SecurityDescriptor  structures.LSAPR_SR_SECURITY_DESCRIPTOR
}

func (*lsarSetSecurityObjectRequest) Opnum() uint16 { return lsarpc.OpnumLsarSetSecurityObject }

// LsarSetSecurityObject calls LsarSetSecurityObject (opnum 4) to set the security
// descriptor of an LSA object ([MS-LSAD] 3.1.4.4.2). securityInformation is a
// SECURITY_INFORMATION bitmask selecting which components of the descriptor to apply.
func LsarSetSecurityObject(rpc ndr.Invoker, objectHandle structures.LSAPR_HANDLE, securityInformation uint32, securityDescriptor structures.LSAPR_SR_SECURITY_DESCRIPTOR) error {
	req := &lsarSetSecurityObjectRequest{
		ObjectHandle:        objectHandle,
		SecurityInformation: ndr.DWORD(securityInformation),
		SecurityDescriptor:  securityDescriptor,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetSecurityObject: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetSecurityObject failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
