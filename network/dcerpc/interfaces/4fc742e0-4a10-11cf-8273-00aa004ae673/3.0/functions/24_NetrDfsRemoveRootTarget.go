package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsRemoveRootTargetRequest carries the [in] parameters of NetrDfsRemoveRootTarget.
type netrDfsRemoveRootTargetRequest struct {
	PDfsPath    *ndr.WSTR `ndr:"unique"`
	PTargetPath *ndr.WSTR `ndr:"unique"`
	Flags       ndr.DWORD
}

func (*netrDfsRemoveRootTargetRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsRemoveRootTarget }

// netrDfsRemoveRootTargetResponse carries the [out] parameters and return value of NetrDfsRemoveRootTarget.
type netrDfsRemoveRootTargetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsRemoveRootTarget calls NetrDfsRemoveRootTarget (opnum 24) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsRemoveRootTarget(rpc ndr.Invoker, pDfsPath *ndr.WSTR, pTargetPath *ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &netrDfsRemoveRootTargetRequest{
		PDfsPath:    pDfsPath,
		PTargetPath: pTargetPath,
		Flags:       flags,
	}
	var resp netrDfsRemoveRootTargetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsRemoveRootTarget: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsRemoveRootTarget failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
