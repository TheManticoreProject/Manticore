package functions

// IDL source: [MS-DFSNM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dfsnm/b471e023-618d-4c48-877f-f30c3005320c
// A fetched copy is kept at ms-dfsnm.idl in the interface directory.

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdfsnm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dfsnm"
)

// netrDfsGetSupportedNamespaceVersionRequest carries the [in] parameters of NetrDfsGetSupportedNamespaceVersion.
type netrDfsGetSupportedNamespaceVersionRequest struct {
	Origin msdfsnm.DFS_NAMESPACE_VERSION_ORIGIN
	PName  *ndr.WSTR `ndr:"unique"`
}

func (*netrDfsGetSupportedNamespaceVersionRequest) Opnum() uint16 {
	return netdfs.OpnumNetrDfsGetSupportedNamespaceVersion
}

// netrDfsGetSupportedNamespaceVersionResponse carries the [out] parameters and return value of NetrDfsGetSupportedNamespaceVersion.
type netrDfsGetSupportedNamespaceVersionResponse struct {
	PVersionInfo msdfsnm.DFS_SUPPORTED_NAMESPACE_VERSION_INFO
	Status       ndr.DWORD `ndr:"retval"`
}

// NetrDfsGetSupportedNamespaceVersion calls NetrDfsGetSupportedNamespaceVersion (opnum 25) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsGetSupportedNamespaceVersion(rpc ndr.Invoker, origin msdfsnm.DFS_NAMESPACE_VERSION_ORIGIN, pName *ndr.WSTR) (PVersionInfo msdfsnm.DFS_SUPPORTED_NAMESPACE_VERSION_INFO, err error) {
	req := &netrDfsGetSupportedNamespaceVersionRequest{
		Origin: origin,
		PName:  pName,
	}
	var resp netrDfsGetSupportedNamespaceVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsGetSupportedNamespaceVersion: %w", err)
		return
	}
	PVersionInfo = resp.PVersionInfo
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsGetSupportedNamespaceVersion failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
