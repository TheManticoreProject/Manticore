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

// schRpcGetNumberOfMissedRunsRequest carries the [in] parameters of SchRpcGetNumberOfMissedRuns.
type schRpcGetNumberOfMissedRunsRequest struct {
	Path ndr.WSTR
}

func (*schRpcGetNumberOfMissedRunsRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcGetNumberOfMissedRuns
}

// schRpcGetNumberOfMissedRunsResponse carries the [out] parameters and return value of SchRpcGetNumberOfMissedRuns.
type schRpcGetNumberOfMissedRunsResponse struct {
	PNumberOfMissedRuns ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// SchRpcGetNumberOfMissedRuns calls SchRpcGetNumberOfMissedRuns (opnum 18) ([MS-TSCH] section 3.2.5.4.19).
func SchRpcGetNumberOfMissedRuns(rpc ndr.Invoker, path ndr.WSTR) (PNumberOfMissedRuns ndr.DWORD, err error) {
	req := &schRpcGetNumberOfMissedRunsRequest{
		Path: path,
	}
	var resp schRpcGetNumberOfMissedRunsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcGetNumberOfMissedRuns: %w", err)
		return
	}
	PNumberOfMissedRuns = resp.PNumberOfMissedRuns
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcGetNumberOfMissedRuns failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
