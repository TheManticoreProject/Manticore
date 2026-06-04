package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarQuerySecurityObjectRequest is the [in] parameter set of LsarQuerySecurityObject: an
// open object handle and the SECURITY_INFORMATION bitmask selecting which parts of the
// descriptor to return.
type lsarQuerySecurityObjectRequest struct {
	ObjectHandle        structures.LSAPR_HANDLE
	SecurityInformation ndr.DWORD
}

func (*lsarQuerySecurityObjectRequest) Opnum() uint16 { return lsarpc.OpnumLsarQuerySecurityObject }

// lsarQuerySecurityObjectResponse is the reply: the [out] self-relative security
// descriptor (a double pointer) followed by the NTSTATUS return value.
type lsarQuerySecurityObjectResponse struct {
	SecurityDescriptor *structures.LSAPR_SR_SECURITY_DESCRIPTOR `ndr:"unique"`
	Status             ndr.DWORD
}

// LsarQuerySecurityObject calls LsarQuerySecurityObject (opnum 3) to retrieve the security
// descriptor of an LSA object ([MS-LSAD] 3.1.4.4.1). securityInformation is a
// SECURITY_INFORMATION bitmask selecting which components of the descriptor to return.
func LsarQuerySecurityObject(rpc *client.Client, objectHandle structures.LSAPR_HANDLE, securityInformation uint32) (*structures.LSAPR_SR_SECURITY_DESCRIPTOR, error) {
	req := &lsarQuerySecurityObjectRequest{
		ObjectHandle:        objectHandle,
		SecurityInformation: ndr.DWORD(securityInformation),
	}
	var resp lsarQuerySecurityObjectResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQuerySecurityObject: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.SecurityDescriptor, fmt.Errorf("LsarQuerySecurityObject failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.SecurityDescriptor, nil
}
