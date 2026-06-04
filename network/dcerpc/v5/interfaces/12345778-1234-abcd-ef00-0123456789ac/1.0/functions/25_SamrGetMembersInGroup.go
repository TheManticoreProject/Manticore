package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrGetMembersInGroupRequest carries the [in] group handle whose members are listed.
type samrGetMembersInGroupRequest struct {
	GroupHandle structures.SAMPR_HANDLE
}

func (*samrGetMembersInGroupRequest) Opnum() uint16 { return samr.OpnumSamrGetMembersInGroup }

// samrGetMembersInGroupResponse is the reply: the [out,unique] members buffer (the relative
// ids and attributes of the group members) and the NTSTATUS.
type samrGetMembersInGroupResponse struct {
	Members *structures.SAMPR_GET_MEMBERS_BUFFER `ndr:"unique"`
	Status  ndr.DWORD                            `ndr:"retval"`
}

// SamrGetMembersInGroup calls SamrGetMembersInGroup (opnum 25), reading the members of a
// group ([MS-SAMR] 3.1.5.8.3).
func SamrGetMembersInGroup(rpc *client.Client, groupHandle structures.SAMPR_HANDLE) (*structures.SAMPR_GET_MEMBERS_BUFFER, error) {
	req := &samrGetMembersInGroupRequest{GroupHandle: groupHandle}
	var resp samrGetMembersInGroupResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrGetMembersInGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Members, fmt.Errorf("SamrGetMembersInGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Members, nil
}
