package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarCloseRequest is the [in,out] parameter of LsarClose: the context handle.
type lsarCloseRequest struct {
	Handle structures.LSAPR_HANDLE
}

func (*lsarCloseRequest) Opnum() uint16 { return lsarpc.OpnumLsarClose }

// LsarClose calls LsarClose (opnum 0) on a handle. On success the server returns a
// zeroed handle, which is returned to the caller.
func LsarClose(rpc ndr.Invoker, handle structures.LSAPR_HANDLE) (structures.LSAPR_HANDLE, error) {
	req := &lsarCloseRequest{Handle: handle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarClose: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarClose failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
