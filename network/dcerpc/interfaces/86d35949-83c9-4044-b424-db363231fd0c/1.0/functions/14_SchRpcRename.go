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

// schRpcRenameRequest carries the [in] parameters of SchRpcRename.
type schRpcRenameRequest struct {
	Path    ndr.WSTR
	NewName ndr.WSTR
	Flags   ndr.DWORD
}

func (*schRpcRenameRequest) Opnum() uint16 { return schrpc.OpnumSchRpcRename }

// schRpcRenameResponse carries the [out] parameters and return value of SchRpcRename.
type schRpcRenameResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcRename calls SchRpcRename (opnum 14) ([MS-TSCH] section 3.2.5.4.15).
func SchRpcRename(rpc ndr.Invoker, path ndr.WSTR, newName ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &schRpcRenameRequest{
		Path:    path,
		NewName: newName,
		Flags:   flags,
	}
	var resp schRpcRenameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcRename: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcRename failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
