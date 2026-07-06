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

// schRpcGetTaskInfoRequest carries the [in] parameters of SchRpcGetTaskInfo.
type schRpcGetTaskInfoRequest struct {
	Path  ndr.WSTR
	Flags ndr.DWORD
}

func (*schRpcGetTaskInfoRequest) Opnum() uint16 { return schrpc.OpnumSchRpcGetTaskInfo }

// schRpcGetTaskInfoResponse carries the [out] parameters and return value of SchRpcGetTaskInfo.
type schRpcGetTaskInfoResponse struct {
	PEnabled ndr.DWORD
	PState   ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// SchRpcGetTaskInfo calls SchRpcGetTaskInfo (opnum 17) ([MS-TSCH] section 3.2.5.4.18).
func SchRpcGetTaskInfo(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD) (PEnabled ndr.DWORD, PState ndr.DWORD, err error) {
	req := &schRpcGetTaskInfoRequest{
		Path:  path,
		Flags: flags,
	}
	var resp schRpcGetTaskInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcGetTaskInfo: %w", err)
		return
	}
	PEnabled = resp.PEnabled
	PState = resp.PState
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcGetTaskInfo failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
