package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetMemberAttributesOfGroupRequest carries the [in] group handle, the relative id of
// the member, and the membership attributes to set for that member.
type samrSetMemberAttributesOfGroupRequest struct {
	GroupHandle mssamr.SAMPR_HANDLE
	MemberId    ndr.DWORD
	Attributes  ndr.DWORD
}

func (*samrSetMemberAttributesOfGroupRequest) Opnum() uint16 {
	return samr.OpnumSamrSetMemberAttributesOfGroup
}

// SamrSetMemberAttributesOfGroup calls SamrSetMemberAttributesOfGroup (opnum 26), setting the
// attributes of a member relationship in a group ([MS-SAMR] 3.1.5.8.4).
func SamrSetMemberAttributesOfGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE, memberId uint32, attributes uint32) error {
	req := &samrSetMemberAttributesOfGroupRequest{
		GroupHandle: groupHandle,
		MemberId:    ndr.DWORD(memberId),
		Attributes:  ndr.DWORD(attributes),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetMemberAttributesOfGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetMemberAttributesOfGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
