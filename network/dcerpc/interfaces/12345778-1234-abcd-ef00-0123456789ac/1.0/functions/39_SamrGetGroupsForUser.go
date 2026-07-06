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

// samrGetGroupsForUserRequest carries the [in] user handle whose group memberships are
// queried.
type samrGetGroupsForUserRequest struct {
	UserHandle mssamr.SAMPR_HANDLE
}

func (*samrGetGroupsForUserRequest) Opnum() uint16 { return samr.OpnumSamrGetGroupsForUser }

// samrGetGroupsForUserResponse is the reply: the [out,unique] buffer of group memberships and
// the NTSTATUS.
type samrGetGroupsForUserResponse struct {
	Groups *mssamr.SAMPR_GET_GROUPS_BUFFER `ndr:"unique"`
	Status ndr.DWORD                       `ndr:"retval"`
}

// SamrGetGroupsForUser calls SamrGetGroupsForUser (opnum 39), returning the union of all
// groups the given user is a member of ([MS-SAMR] 3.1.5.9.1).
func SamrGetGroupsForUser(rpc ndr.Invoker, userHandle mssamr.SAMPR_HANDLE) (*mssamr.SAMPR_GET_GROUPS_BUFFER, error) {
	req := &samrGetGroupsForUserRequest{UserHandle: userHandle}
	var resp samrGetGroupsForUserResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrGetGroupsForUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Groups, fmt.Errorf("SamrGetGroupsForUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Groups, nil
}
