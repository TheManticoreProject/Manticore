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

// netrDfsAddFtRootRequest carries the [in] parameters of NetrDfsAddFtRoot.
type netrDfsAddFtRootRequest struct {
	ServerName ndr.WSTR
	DcName     ndr.WSTR
	RootShare  ndr.WSTR
	FtDfsName  ndr.WSTR
	Comment    ndr.WSTR
	ConfigDN   ndr.WSTR
	NewFtDfs   bool
	ApiFlags   ndr.DWORD
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
}

func (*netrDfsAddFtRootRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsAddFtRoot }

// netrDfsAddFtRootResponse carries the [out] parameters and return value of NetrDfsAddFtRoot.
type netrDfsAddFtRootResponse struct {
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// NetrDfsAddFtRoot calls NetrDfsAddFtRoot (opnum 10) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAddFtRoot(rpc ndr.Invoker, serverName ndr.WSTR, dcName ndr.WSTR, rootShare ndr.WSTR, ftDfsName ndr.WSTR, comment ndr.WSTR, configDN ndr.WSTR, newFtDfs bool, apiFlags ndr.DWORD, ppRootList *msdfsnm.DFSM_ROOT_LIST) (PpRootList *msdfsnm.DFSM_ROOT_LIST, err error) {
	req := &netrDfsAddFtRootRequest{
		ServerName: serverName,
		DcName:     dcName,
		RootShare:  rootShare,
		FtDfsName:  ftDfsName,
		Comment:    comment,
		ConfigDN:   configDN,
		NewFtDfs:   newFtDfs,
		ApiFlags:   apiFlags,
		PpRootList: ppRootList,
	}
	var resp netrDfsAddFtRootResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAddFtRoot: %w", err)
		return
	}
	PpRootList = resp.PpRootList
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAddFtRoot failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
