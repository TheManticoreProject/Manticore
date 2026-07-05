package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrDeleteGroupRequest carries the [in,out] SAMPR_HANDLE of the group to delete. On
// success the server returns it zeroed via the shared handleResponse.
type samrDeleteGroupRequest struct {
	GroupHandle mssamr.SAMPR_HANDLE
}

func (*samrDeleteGroupRequest) Opnum() uint16 { return samr.OpnumSamrDeleteGroup }

// SamrDeleteGroup calls SamrDeleteGroup (opnum 23), removing a group object from the database
// and returning the (now zeroed) handle ([MS-SAMR] 3.1.5.7.1).
func SamrDeleteGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE) (mssamr.SAMPR_HANDLE, error) {
	req := &samrDeleteGroupRequest{GroupHandle: groupHandle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return groupHandle, fmt.Errorf("SamrDeleteGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrDeleteGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
