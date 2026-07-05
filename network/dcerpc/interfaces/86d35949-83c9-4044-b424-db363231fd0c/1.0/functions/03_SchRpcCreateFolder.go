package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// schRpcCreateFolderRequest carries the [in] parameters of SchRpcCreateFolder.
type schRpcCreateFolderRequest struct {
	Path  ndr.WSTR
	Sddl  *ndr.WSTR `ndr:"unique"`
	Flags ndr.DWORD
}

func (*schRpcCreateFolderRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcCreateFolder
}

// schRpcCreateFolderResponse carries the [out] parameters and return value of SchRpcCreateFolder.
type schRpcCreateFolderResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcCreateFolder calls SchRpcCreateFolder (opnum 3) ([MS-TSCH] section 3.2.5.4.4).
func SchRpcCreateFolder(rpc ndr.Invoker, path ndr.WSTR, sddl *ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &schRpcCreateFolderRequest{
		Path:  path,
		Sddl:  sddl,
		Flags: flags,
	}
	var resp schRpcCreateFolderResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcCreateFolder: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcCreateFolder failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
