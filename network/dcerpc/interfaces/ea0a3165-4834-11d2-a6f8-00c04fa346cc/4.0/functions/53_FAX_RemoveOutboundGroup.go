package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_RemoveOutboundGroupRequest carries the [in] parameters of FAX_RemoveOutboundGroup.
type fAX_RemoveOutboundGroupRequest struct {
	LpwstrGroupName ndr.WSTR
}

func (*fAX_RemoveOutboundGroupRequest) Opnum() uint16 { return fax.OpnumFAX_RemoveOutboundGroup }

// fAX_RemoveOutboundGroupResponse carries the [out] parameters and return value of FAX_RemoveOutboundGroup.
type fAX_RemoveOutboundGroupResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_RemoveOutboundGroup calls FAX_RemoveOutboundGroup (opnum 53) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_RemoveOutboundGroup(rpc ndr.Invoker, lpwstrGroupName ndr.WSTR) (err error) {
	req := &fAX_RemoveOutboundGroupRequest{
		LpwstrGroupName: lpwstrGroupName,
	}
	var resp fAX_RemoveOutboundGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_RemoveOutboundGroup: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_RemoveOutboundGroup failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
