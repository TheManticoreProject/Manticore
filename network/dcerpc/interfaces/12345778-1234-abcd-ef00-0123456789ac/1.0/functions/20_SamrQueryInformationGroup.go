package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrQueryInformationGroupRequest carries the [in] group handle and the information class
// selecting which arm of the returned union is populated.
type samrQueryInformationGroupRequest struct {
	GroupHandle           structures.SAMPR_HANDLE
	GroupInformationClass structures.GROUP_INFORMATION_CLASS
}

func (*samrQueryInformationGroupRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationGroup
}

// samrQueryInformationGroupResponse is the reply: the [out,switch_is,unique] group info
// buffer (carrying its own discriminant) and the NTSTATUS.
type samrQueryInformationGroupResponse struct {
	Buffer *structures.SAMPR_GROUP_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                           `ndr:"retval"`
}

// SamrQueryInformationGroup calls SamrQueryInformationGroup (opnum 20), retrieving
// attributes of a group object ([MS-SAMR] 3.1.5.5.3).
func SamrQueryInformationGroup(rpc ndr.Invoker, groupHandle structures.SAMPR_HANDLE, groupInformationClass structures.GROUP_INFORMATION_CLASS) (*structures.SAMPR_GROUP_INFO_BUFFER, error) {
	req := &samrQueryInformationGroupRequest{
		GroupHandle:           groupHandle,
		GroupInformationClass: groupInformationClass,
	}
	var resp samrQueryInformationGroupResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}
