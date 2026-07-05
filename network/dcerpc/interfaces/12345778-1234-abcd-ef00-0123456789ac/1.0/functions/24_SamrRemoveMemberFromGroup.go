package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrRemoveMemberFromGroupRequest carries the [in] group handle and the relative id of the
// member to remove.
type samrRemoveMemberFromGroupRequest struct {
	GroupHandle mssamr.SAMPR_HANDLE
	MemberId    ndr.DWORD
}

func (*samrRemoveMemberFromGroupRequest) Opnum() uint16 {
	return samr.OpnumSamrRemoveMemberFromGroup
}

// SamrRemoveMemberFromGroup calls SamrRemoveMemberFromGroup (opnum 24), removing a user from
// the membership of a group ([MS-SAMR] 3.1.5.8.2).
func SamrRemoveMemberFromGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE, memberId uint32) error {
	req := &samrRemoveMemberFromGroupRequest{
		GroupHandle: groupHandle,
		MemberId:    ndr.DWORD(memberId),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrRemoveMemberFromGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrRemoveMemberFromGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
