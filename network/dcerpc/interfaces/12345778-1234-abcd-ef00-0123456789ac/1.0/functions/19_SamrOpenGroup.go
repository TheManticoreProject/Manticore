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

// samrOpenGroupRequest carries the [in] parameters of SamrOpenGroup: the domain handle, the
// desired access mask, and the relative id of the group to open.
type samrOpenGroupRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	GroupId       ndr.DWORD
}

func (*samrOpenGroupRequest) Opnum() uint16 { return samr.OpnumSamrOpenGroup }

// SamrOpenGroup calls SamrOpenGroup (opnum 19), obtaining a handle to a group object
// ([MS-SAMR] 3.1.5.1.5).
func SamrOpenGroup(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, desiredAccess uint32, groupId uint32) (mssamr.SAMPR_HANDLE, error) {
	req := &samrOpenGroupRequest{
		DomainHandle:  domainHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		GroupId:       ndr.DWORD(groupId),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
