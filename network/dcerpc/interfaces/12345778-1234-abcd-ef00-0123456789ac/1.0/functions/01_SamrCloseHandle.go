package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrCloseHandleRequest carries the [in,out] SAMPR_HANDLE to be closed. On
// success the server returns it zeroed via the shared handleResponse.
type samrCloseHandleRequest struct {
	SamHandle mssamr.SAMPR_HANDLE
}

func (*samrCloseHandleRequest) Opnum() uint16 { return samr.OpnumSamrCloseHandle }

// SamrCloseHandle calls SamrCloseHandle (opnum 1), releasing server resources for
// the supplied handle and returning the (now zeroed) handle ([MS-SAMR] 3.1.5.13.1).
func SamrCloseHandle(rpc ndr.Invoker, handle mssamr.SAMPR_HANDLE) (mssamr.SAMPR_HANDLE, error) {
	req := &samrCloseHandleRequest{SamHandle: handle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return handle, fmt.Errorf("SamrCloseHandle: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrCloseHandle failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
