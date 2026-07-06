package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACCloseCursorRequest carries the [in] parameters of rpc_ACCloseCursor.
type rpc_ACCloseCursorRequest struct {
	HQueue  msmqmp.RPC_QUEUE_HANDLE
	HCursor ndr.DWORD
}

func (*rpc_ACCloseCursorRequest) Opnum() uint16 { return qmcomm.Opnumrpc_ACCloseCursor }

// rpc_ACCloseCursorResponse carries the [out] parameters and return value of rpc_ACCloseCursor.
type rpc_ACCloseCursorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Rpc_ACCloseCursor calls rpc_ACCloseCursor (opnum 22) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_ACCloseCursor(rpc ndr.Invoker, hQueue msmqmp.RPC_QUEUE_HANDLE, hCursor ndr.DWORD) (err error) {
	req := &rpc_ACCloseCursorRequest{
		HQueue:  hQueue,
		HCursor: hCursor,
	}
	var resp rpc_ACCloseCursorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACCloseCursor: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_ACCloseCursor failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
