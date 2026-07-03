package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACSetCursorPropertiesRequest carries the [in] parameters of rpc_ACSetCursorProperties.
type rpc_ACSetCursorPropertiesRequest struct {
	HProxy        msmqmp.RPC_QUEUE_HANDLE
	HCursor       ndr.DWORD
	HRemoteCursor ndr.DWORD
}

func (*rpc_ACSetCursorPropertiesRequest) Opnum() uint16 { return qmcomm.Opnumrpc_ACSetCursorProperties }

// rpc_ACSetCursorPropertiesResponse carries the [out] parameters and return value of rpc_ACSetCursorProperties.
type rpc_ACSetCursorPropertiesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Rpc_ACSetCursorProperties calls rpc_ACSetCursorProperties (opnum 23) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_ACSetCursorProperties(rpc ndr.Invoker, hProxy msmqmp.RPC_QUEUE_HANDLE, hCursor ndr.DWORD, hRemoteCursor ndr.DWORD) (err error) {
	req := &rpc_ACSetCursorPropertiesRequest{
		HProxy:        hProxy,
		HCursor:       hCursor,
		HRemoteCursor: hRemoteCursor,
	}
	var resp rpc_ACSetCursorPropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACSetCursorProperties: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_ACSetCursorProperties failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
