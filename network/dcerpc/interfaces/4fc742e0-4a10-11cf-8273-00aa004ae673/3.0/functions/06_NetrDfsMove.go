package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsMoveRequest carries the [in] parameters of NetrDfsMove.
type netrDfsMoveRequest struct {
	DfsEntryPath    ndr.WSTR
	NewDfsEntryPath ndr.WSTR
	Flags           ndr.DWORD
}

func (*netrDfsMoveRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsMove }

// netrDfsMoveResponse carries the [out] parameters and return value of NetrDfsMove.
type netrDfsMoveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsMove calls NetrDfsMove (opnum 6) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsMove(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, newDfsEntryPath ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &netrDfsMoveRequest{
		DfsEntryPath:    dfsEntryPath,
		NewDfsEntryPath: newDfsEntryPath,
		Flags:           flags,
	}
	var resp netrDfsMoveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsMove: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsMove failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
