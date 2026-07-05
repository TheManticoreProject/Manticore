package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrOpenUserRequest carries the [in] parameters of SamrOpenUser: the domain handle, the
// desired access mask, and the relative id of the user to open.
type samrOpenUserRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	UserId        ndr.DWORD
}

func (*samrOpenUserRequest) Opnum() uint16 { return samr.OpnumSamrOpenUser }

// SamrOpenUser calls SamrOpenUser (opnum 34), obtaining a handle to a user object given its
// relative id within the domain ([MS-SAMR] 3.1.5.1.9).
func SamrOpenUser(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, desiredAccess uint32, userId uint32) (mssamr.SAMPR_HANDLE, error) {
	req := &samrOpenUserRequest{
		DomainHandle:  domainHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		UserId:        ndr.DWORD(userId),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
