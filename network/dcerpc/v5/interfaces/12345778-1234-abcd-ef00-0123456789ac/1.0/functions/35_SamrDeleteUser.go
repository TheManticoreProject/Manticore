package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrDeleteUserRequest carries the [in,out] SAMPR_HANDLE of the user to delete. On success
// the server returns it zeroed via the shared handleResponse.
type samrDeleteUserRequest struct {
	UserHandle structures.SAMPR_HANDLE
}

func (*samrDeleteUserRequest) Opnum() uint16 { return samr.OpnumSamrDeleteUser }

// SamrDeleteUser calls SamrDeleteUser (opnum 35), removing the user object referenced by the
// handle and returning the (now zeroed) handle ([MS-SAMR] 3.1.5.7.3).
func SamrDeleteUser(rpc *client.Client, userHandle structures.SAMPR_HANDLE) (structures.SAMPR_HANDLE, error) {
	req := &samrDeleteUserRequest{UserHandle: userHandle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return userHandle, fmt.Errorf("SamrDeleteUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrDeleteUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
