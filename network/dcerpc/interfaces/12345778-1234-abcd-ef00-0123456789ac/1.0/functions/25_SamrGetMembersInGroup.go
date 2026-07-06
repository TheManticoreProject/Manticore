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

// samrGetMembersInGroupRequest carries the [in] group handle whose members are listed.
type samrGetMembersInGroupRequest struct {
	GroupHandle mssamr.SAMPR_HANDLE
}

func (*samrGetMembersInGroupRequest) Opnum() uint16 { return samr.OpnumSamrGetMembersInGroup }

// samrGetMembersInGroupResponse is the reply: the [out,unique] members buffer (the relative
// ids and attributes of the group members) and the NTSTATUS.
type samrGetMembersInGroupResponse struct {
	Members *mssamr.SAMPR_GET_MEMBERS_BUFFER `ndr:"unique"`
	Status  ndr.DWORD                        `ndr:"retval"`
}

// SamrGetMembersInGroup calls SamrGetMembersInGroup (opnum 25), reading the members of a
// group ([MS-SAMR] 3.1.5.8.3).
func SamrGetMembersInGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE) (*mssamr.SAMPR_GET_MEMBERS_BUFFER, error) {
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
