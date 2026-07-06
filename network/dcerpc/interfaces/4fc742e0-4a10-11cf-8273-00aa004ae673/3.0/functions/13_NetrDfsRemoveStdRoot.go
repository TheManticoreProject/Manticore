package functions

// IDL source: [MS-DFSNM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dfsnm/b471e023-618d-4c48-877f-f30c3005320c
// A fetched copy is kept at ms-dfsnm.idl in the interface directory.

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsRemoveStdRootRequest carries the [in] parameters of NetrDfsRemoveStdRoot.
type netrDfsRemoveStdRootRequest struct {
	ServerName ndr.WSTR
	RootShare  ndr.WSTR
	ApiFlags   ndr.DWORD
}

func (*netrDfsRemoveStdRootRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsRemoveStdRoot }

// netrDfsRemoveStdRootResponse carries the [out] parameters and return value of NetrDfsRemoveStdRoot.
type netrDfsRemoveStdRootResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsRemoveStdRoot calls NetrDfsRemoveStdRoot (opnum 13) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsRemoveStdRoot(rpc ndr.Invoker, serverName ndr.WSTR, rootShare ndr.WSTR, apiFlags ndr.DWORD) (err error) {
	req := &netrDfsRemoveStdRootRequest{
		ServerName: serverName,
		RootShare:  rootShare,
		ApiFlags:   apiFlags,
	}
	var resp netrDfsRemoveStdRootResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsRemoveStdRoot: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsRemoveStdRoot failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
