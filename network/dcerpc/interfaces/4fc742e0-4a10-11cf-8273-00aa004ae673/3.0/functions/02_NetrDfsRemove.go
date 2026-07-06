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

// netrDfsRemoveRequest carries the [in] parameters of NetrDfsRemove.
type netrDfsRemoveRequest struct {
	DfsEntryPath ndr.WSTR
	ServerName   *ndr.WSTR `ndr:"unique"`
	ShareName    *ndr.WSTR `ndr:"unique"`
}

func (*netrDfsRemoveRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsRemove }

// netrDfsRemoveResponse carries the [out] parameters and return value of NetrDfsRemove.
type netrDfsRemoveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsRemove calls NetrDfsRemove (opnum 2) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsRemove(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, serverName *ndr.WSTR, shareName *ndr.WSTR) (err error) {
	req := &netrDfsRemoveRequest{
		DfsEntryPath: dfsEntryPath,
		ServerName:   serverName,
		ShareName:    shareName,
	}
	var resp netrDfsRemoveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsRemove: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsRemove failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
