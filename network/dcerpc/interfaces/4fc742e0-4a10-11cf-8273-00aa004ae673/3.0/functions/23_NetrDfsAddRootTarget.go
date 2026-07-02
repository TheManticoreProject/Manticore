package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsAddRootTargetRequest carries the [in] parameters of NetrDfsAddRootTarget.
type netrDfsAddRootTargetRequest struct {
	PDfsPath     *ndr.WSTR `ndr:"unique"`
	PTargetPath  *ndr.WSTR `ndr:"unique"`
	MajorVersion ndr.DWORD
	PComment     *ndr.WSTR `ndr:"unique"`
	NewNamespace bool
	Flags        ndr.DWORD
}

func (*netrDfsAddRootTargetRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsAddRootTarget }

// netrDfsAddRootTargetResponse carries the [out] parameters and return value of NetrDfsAddRootTarget.
type netrDfsAddRootTargetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsAddRootTarget calls NetrDfsAddRootTarget (opnum 23) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAddRootTarget(rpc ndr.Invoker, pDfsPath *ndr.WSTR, pTargetPath *ndr.WSTR, majorVersion ndr.DWORD, pComment *ndr.WSTR, newNamespace bool, flags ndr.DWORD) (err error) {
	req := &netrDfsAddRootTargetRequest{
		PDfsPath:     pDfsPath,
		PTargetPath:  pTargetPath,
		MajorVersion: majorVersion,
		PComment:     pComment,
		NewNamespace: newNamespace,
		Flags:        flags,
	}
	var resp netrDfsAddRootTargetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAddRootTarget: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAddRootTarget failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
