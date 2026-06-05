package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrCloseHandleRequest carries the [in,out] SAMPR_HANDLE to be closed. On
// success the server returns it zeroed via the shared handleResponse.
type samrCloseHandleRequest struct {
	SamHandle structures.SAMPR_HANDLE
}

func (*samrCloseHandleRequest) Opnum() uint16 { return samr.OpnumSamrCloseHandle }

// SamrCloseHandle calls SamrCloseHandle (opnum 1), releasing server resources for
// the supplied handle and returning the (now zeroed) handle ([MS-SAMR] 3.1.5.13.1).
func SamrCloseHandle(rpc ndr.Invoker, handle structures.SAMPR_HANDLE) (structures.SAMPR_HANDLE, error) {
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
