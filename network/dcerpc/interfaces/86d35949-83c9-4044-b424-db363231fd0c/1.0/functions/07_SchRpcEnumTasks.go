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

// schRpcEnumTasksRequest carries the [in] parameters of SchRpcEnumTasks.
type schRpcEnumTasksRequest struct {
	Path       ndr.WSTR
	Flags      ndr.DWORD
	StartIndex ndr.DWORD
	CRequested ndr.DWORD
}

func (*schRpcEnumTasksRequest) Opnum() uint16 { return schrpc.OpnumSchRpcEnumTasks }

// schRpcEnumTasksResponse carries the [out] parameters and return value of SchRpcEnumTasks.
type schRpcEnumTasksResponse struct {
	StartIndex ndr.DWORD
	PcNames    ndr.DWORD
	PNames     []*ndr.WSTR `ndr:"unique,size_is=PcNames"`
	Status     ndr.DWORD   `ndr:"retval"`
}

// SchRpcEnumTasks calls SchRpcEnumTasks (opnum 7) ([MS-TSCH] section 3.2.5.4.8).
func SchRpcEnumTasks(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD, startIndex ndr.DWORD, cRequested ndr.DWORD) (StartIndex ndr.DWORD, PcNames ndr.DWORD, PNames []*ndr.WSTR, err error) {
	req := &schRpcEnumTasksRequest{
		Path:       path,
		Flags:      flags,
		StartIndex: startIndex,
		CRequested: cRequested,
	}
	var resp schRpcEnumTasksResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcEnumTasks: %w", err)
		return
	}
	StartIndex = resp.StartIndex
	PcNames = resp.PcNames
	PNames = resp.PNames
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcEnumTasks failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
