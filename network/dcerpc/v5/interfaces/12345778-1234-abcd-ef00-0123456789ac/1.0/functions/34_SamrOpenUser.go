package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrOpenUserRequest carries the [in] parameters of SamrOpenUser: the domain handle, the
// desired access mask, and the relative id of the user to open.
type samrOpenUserRequest struct {
	DomainHandle  structures.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	UserId        ndr.DWORD
}

func (*samrOpenUserRequest) Opnum() uint16 { return samr.OpnumSamrOpenUser }

// SamrOpenUser calls SamrOpenUser (opnum 34), obtaining a handle to a user object given its
// relative id within the domain ([MS-SAMR] 3.1.5.1.9).
func SamrOpenUser(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, desiredAccess uint32, userId uint32) (structures.SAMPR_HANDLE, error) {
	req := &samrOpenUserRequest{
		DomainHandle:  domainHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		UserId:        ndr.DWORD(userId),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
