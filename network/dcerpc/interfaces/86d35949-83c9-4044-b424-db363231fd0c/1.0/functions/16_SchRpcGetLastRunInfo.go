package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// schRpcGetLastRunInfoRequest carries the [in] parameters of SchRpcGetLastRunInfo.
type schRpcGetLastRunInfoRequest struct {
	Path ndr.WSTR
}

func (*schRpcGetLastRunInfoRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcGetLastRunInfo
}

// schRpcGetLastRunInfoResponse carries the [out] parameters and return value of SchRpcGetLastRunInfo.
type schRpcGetLastRunInfoResponse struct {
	PLastRuntime    msdtyp.SYSTEMTIME
	PLastReturnCode ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// SchRpcGetLastRunInfo calls SchRpcGetLastRunInfo (opnum 16) ([MS-TSCH] section 3.2.5.4.17).
func SchRpcGetLastRunInfo(rpc ndr.Invoker, path ndr.WSTR) (PLastRuntime msdtyp.SYSTEMTIME, PLastReturnCode ndr.DWORD, err error) {
	req := &schRpcGetLastRunInfoRequest{
		Path: path,
	}
	var resp schRpcGetLastRunInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcGetLastRunInfo: %w", err)
		return
	}
	PLastRuntime = resp.PLastRuntime
	PLastReturnCode = resp.PLastReturnCode
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcGetLastRunInfo failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
