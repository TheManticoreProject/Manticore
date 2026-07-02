package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SetOutboundGroupRequest carries the [in] parameters of FAX_SetOutboundGroup.
type fAX_SetOutboundGroupRequest struct {
	PGroup msfax.RPC_FAX_OUTBOUND_ROUTING_GROUPW
}

func (*fAX_SetOutboundGroupRequest) Opnum() uint16 { return fax.OpnumFAX_SetOutboundGroup }

// fAX_SetOutboundGroupResponse carries the [out] parameters and return value of FAX_SetOutboundGroup.
type fAX_SetOutboundGroupResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetOutboundGroup calls FAX_SetOutboundGroup (opnum 52) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetOutboundGroup(rpc ndr.Invoker, pGroup msfax.RPC_FAX_OUTBOUND_ROUTING_GROUPW) (err error) {
	req := &fAX_SetOutboundGroupRequest{
		PGroup: pGroup,
	}
	var resp fAX_SetOutboundGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetOutboundGroup: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetOutboundGroup failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
