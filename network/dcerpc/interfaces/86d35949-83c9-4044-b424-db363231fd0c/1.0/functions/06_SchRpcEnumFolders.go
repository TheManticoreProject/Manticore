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

// schRpcEnumFoldersRequest carries the [in] parameters of SchRpcEnumFolders.
type schRpcEnumFoldersRequest struct {
	Path        ndr.WSTR
	Flags       ndr.DWORD
	PStartIndex ndr.DWORD
	CRequested  ndr.DWORD
}

func (*schRpcEnumFoldersRequest) Opnum() uint16 { return schrpc.OpnumSchRpcEnumFolders }

// schRpcEnumFoldersResponse carries the [out] parameters and return value of SchRpcEnumFolders.
type schRpcEnumFoldersResponse struct {
	PStartIndex ndr.DWORD
	PcNames     ndr.DWORD
	PNames      []*ndr.WSTR `ndr:"unique,size_is=PcNames"`
	Status      ndr.DWORD   `ndr:"retval"`
}

// SchRpcEnumFolders calls SchRpcEnumFolders (opnum 6) ([MS-TSCH] section 3.2.5.4.7).
func SchRpcEnumFolders(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD, pStartIndex ndr.DWORD, cRequested ndr.DWORD) (PStartIndex ndr.DWORD, PcNames ndr.DWORD, PNames []*ndr.WSTR, err error) {
	req := &schRpcEnumFoldersRequest{
		Path:        path,
		Flags:       flags,
		PStartIndex: pStartIndex,
		CRequested:  cRequested,
	}
	var resp schRpcEnumFoldersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcEnumFolders: %w", err)
		return
	}
	PStartIndex = resp.PStartIndex
	PcNames = resp.PcNames
	PNames = resp.PNames
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcEnumFolders failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
