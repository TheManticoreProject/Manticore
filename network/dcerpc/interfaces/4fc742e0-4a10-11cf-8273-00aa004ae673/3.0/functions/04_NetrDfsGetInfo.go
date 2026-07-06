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

// netrDfsGetInfoRequest carries the [in] parameters of NetrDfsGetInfo.
type netrDfsGetInfoRequest struct {
	DfsEntryPath ndr.WSTR
	ServerName   *ndr.WSTR `ndr:"unique"`
	ShareName    *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
}

func (*netrDfsGetInfoRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsGetInfo }

// netrDfsGetInfoResponse carries the [out] parameters and return value of NetrDfsGetInfo.
type netrDfsGetInfoResponse struct {
	DfsInfo msdfsnm.DFS_INFO_STRUCT
	Status  ndr.DWORD `ndr:"retval"`
}

// NetrDfsGetInfo calls NetrDfsGetInfo (opnum 4) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsGetInfo(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, serverName *ndr.WSTR, shareName *ndr.WSTR, level ndr.DWORD) (DfsInfo msdfsnm.DFS_INFO_STRUCT, err error) {
	req := &netrDfsGetInfoRequest{
		DfsEntryPath: dfsEntryPath,
		ServerName:   serverName,
		ShareName:    shareName,
		Level:        level,
	}
	var resp netrDfsGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsGetInfo: %w", err)
		return
	}
	DfsInfo = resp.DfsInfo
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsGetInfo failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
