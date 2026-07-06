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

// samrAddMemberToGroupRequest carries the [in] group handle, the relative id of the member
// to add, and the membership attributes to assign.
type samrAddMemberToGroupRequest struct {
	GroupHandle mssamr.SAMPR_HANDLE
	MemberId    ndr.DWORD
	Attributes  ndr.DWORD
}

func (*samrAddMemberToGroupRequest) Opnum() uint16 { return samr.OpnumSamrAddMemberToGroup }

// SamrAddMemberToGroup calls SamrAddMemberToGroup (opnum 22), adding a user as a member of a
// group ([MS-SAMR] 3.1.5.8.1).
func SamrAddMemberToGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE, memberId uint32, attributes uint32) error {
	req := &samrAddMemberToGroupRequest{
		GroupHandle: groupHandle,
		MemberId:    ndr.DWORD(memberId),
		Attributes:  ndr.DWORD(attributes),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrAddMemberToGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrAddMemberToGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
