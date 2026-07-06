package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// schRpcStopRequest carries the [in] parameters of SchRpcStop.
type schRpcStopRequest struct {
	Path  *ndr.WSTR `ndr:"unique"`
	Flags ndr.DWORD
}

func (*schRpcStopRequest) Opnum() uint16 { return schrpc.OpnumSchRpcStop }

// schRpcStopResponse carries the [out] parameters and return value of SchRpcStop.
type schRpcStopResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcStop calls SchRpcStop (opnum 11) ([MS-TSCH] section 3.2.5.4.12).
func SchRpcStop(rpc ndr.Invoker, path *ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &schRpcStopRequest{
		Path:  path,
		Flags: flags,
	}
	var resp schRpcStopResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcStop: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcStop failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
