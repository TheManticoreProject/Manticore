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

// netrDfsRemoveFtRootRequest carries the [in] parameters of NetrDfsRemoveFtRoot.
type netrDfsRemoveFtRootRequest struct {
	ServerName ndr.WSTR
	DcName     ndr.WSTR
	RootShare  ndr.WSTR
	FtDfsName  ndr.WSTR
	ApiFlags   ndr.DWORD
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
}

func (*netrDfsRemoveFtRootRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsRemoveFtRoot }

// netrDfsRemoveFtRootResponse carries the [out] parameters and return value of NetrDfsRemoveFtRoot.
type netrDfsRemoveFtRootResponse struct {
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// NetrDfsRemoveFtRoot calls NetrDfsRemoveFtRoot (opnum 11) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsRemoveFtRoot(rpc ndr.Invoker, serverName ndr.WSTR, dcName ndr.WSTR, rootShare ndr.WSTR, ftDfsName ndr.WSTR, apiFlags ndr.DWORD, ppRootList *msdfsnm.DFSM_ROOT_LIST) (PpRootList *msdfsnm.DFSM_ROOT_LIST, err error) {
	req := &netrDfsRemoveFtRootRequest{
		ServerName: serverName,
		DcName:     dcName,
		RootShare:  rootShare,
		FtDfsName:  ftDfsName,
		ApiFlags:   apiFlags,
		PpRootList: ppRootList,
	}
	var resp netrDfsRemoveFtRootResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsRemoveFtRoot: %w", err)
		return
	}
	PpRootList = resp.PpRootList
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsRemoveFtRoot failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
