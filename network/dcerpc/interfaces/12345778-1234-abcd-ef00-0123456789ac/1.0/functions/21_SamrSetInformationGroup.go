package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetInformationGroupRequest carries the [in] group handle, the information class, and
// the [in,switch_is] group info buffer whose populated arm matches that class.
type samrSetInformationGroupRequest struct {
	GroupHandle           mssamr.SAMPR_HANDLE
	GroupInformationClass mssamr.GROUP_INFORMATION_CLASS `ndr:"enum"`
	Buffer                mssamr.SAMPR_GROUP_INFO_BUFFER
}

func (*samrSetInformationGroupRequest) Opnum() uint16 { return samr.OpnumSamrSetInformationGroup }

// SamrSetInformationGroup calls SamrSetInformationGroup (opnum 21), updating attributes of a
// group object ([MS-SAMR] 3.1.5.6.2).
func SamrSetInformationGroup(rpc ndr.Invoker, groupHandle mssamr.SAMPR_HANDLE, groupInformationClass mssamr.GROUP_INFORMATION_CLASS, buffer mssamr.SAMPR_GROUP_INFO_BUFFER) error {
	req := &samrSetInformationGroupRequest{
		GroupHandle:           groupHandle,
		GroupInformationClass: groupInformationClass,
		Buffer:                buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetInformationGroup: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetInformationGroup failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
