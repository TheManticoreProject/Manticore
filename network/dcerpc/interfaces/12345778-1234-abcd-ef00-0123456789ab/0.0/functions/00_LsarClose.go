package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarCloseRequest is the [in,out] parameter of LsarClose: the context handle.
type lsarCloseRequest struct {
	Handle mslsad.LSAPR_HANDLE
}

func (*lsarCloseRequest) Opnum() uint16 { return lsarpc.OpnumLsarClose }

// LsarClose calls LsarClose (opnum 0) on a handle. On success the server returns a
// zeroed handle, which is returned to the caller.
func LsarClose(rpc ndr.Invoker, handle mslsad.LSAPR_HANDLE) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarCloseRequest{Handle: handle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarClose: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarClose failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
