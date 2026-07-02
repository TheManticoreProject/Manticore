package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarDeleteObjectRequest is the [in,out] parameter of LsarDeleteObject: the context
// handle of the object (account, secret, or trusted domain) to delete.
type lsarDeleteObjectRequest struct {
	Handle mslsad.LSAPR_HANDLE
}

func (*lsarDeleteObjectRequest) Opnum() uint16 { return lsarpc.OpnumLsarDeleteObject }

// LsarDeleteObject calls LsarDeleteObject (opnum 34), deleting the object referenced by
// the handle. On success the server returns a zeroed handle, which is returned to the
// caller.
func LsarDeleteObject(rpc ndr.Invoker, handle mslsad.LSAPR_HANDLE) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarDeleteObjectRequest{Handle: handle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarDeleteObject: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarDeleteObject failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
