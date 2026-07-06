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

// schRpcEnableTaskRequest carries the [in] parameters of SchRpcEnableTask.
type schRpcEnableTaskRequest struct {
	Path    ndr.WSTR
	Enabled ndr.DWORD
}

func (*schRpcEnableTaskRequest) Opnum() uint16 { return schrpc.OpnumSchRpcEnableTask }

// schRpcEnableTaskResponse carries the [out] parameters and return value of SchRpcEnableTask.
type schRpcEnableTaskResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcEnableTask calls SchRpcEnableTask (opnum 19) ([MS-TSCH] section 3.2.5.4.20).
func SchRpcEnableTask(rpc ndr.Invoker, path ndr.WSTR, enabled ndr.DWORD) (err error) {
	req := &schRpcEnableTaskRequest{
		Path:    path,
		Enabled: enabled,
	}
	var resp schRpcEnableTaskResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcEnableTask: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcEnableTask failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
