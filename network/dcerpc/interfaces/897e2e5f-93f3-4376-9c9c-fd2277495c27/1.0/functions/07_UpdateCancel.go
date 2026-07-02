package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// updateCancelRequest carries the [in] parameters of UpdateCancel.
type updateCancelRequest struct {
	ConnectionId msfrs2.FRS_CONNECTION_ID
	CancelData   msfrs2.FRS_UPDATE_CANCEL_DATA
}

func (*updateCancelRequest) Opnum() uint16 { return FrsTransport.OpnumUpdateCancel }

// updateCancelResponse carries the [out] parameters and return value of UpdateCancel.
type updateCancelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// UpdateCancel calls UpdateCancel (opnum 7) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func UpdateCancel(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID, cancelData msfrs2.FRS_UPDATE_CANCEL_DATA) (err error) {
	req := &updateCancelRequest{
		ConnectionId: connectionId,
		CancelData:   cancelData,
	}
	var resp updateCancelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("UpdateCancel: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("UpdateCancel failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
