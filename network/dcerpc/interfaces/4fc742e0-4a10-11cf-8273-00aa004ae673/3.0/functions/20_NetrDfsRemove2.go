package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdfsnm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dfsnm"
)

// netrDfsRemove2Request carries the [in] parameters of NetrDfsRemove2.
type netrDfsRemove2Request struct {
	DfsEntryPath ndr.WSTR
	DcName       ndr.WSTR
	ServerName   *ndr.WSTR               `ndr:"unique"`
	ShareName    *ndr.WSTR               `ndr:"unique"`
	PpRootList   *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
}

func (*netrDfsRemove2Request) Opnum() uint16 { return netdfs.OpnumNetrDfsRemove2 }

// netrDfsRemove2Response carries the [out] parameters and return value of NetrDfsRemove2.
type netrDfsRemove2Response struct {
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// NetrDfsRemove2 calls NetrDfsRemove2 (opnum 20) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsRemove2(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, dcName ndr.WSTR, serverName *ndr.WSTR, shareName *ndr.WSTR, ppRootList *msdfsnm.DFSM_ROOT_LIST) (PpRootList *msdfsnm.DFSM_ROOT_LIST, err error) {
	req := &netrDfsRemove2Request{
		DfsEntryPath: dfsEntryPath,
		DcName:       dcName,
		ServerName:   serverName,
		ShareName:    shareName,
		PpRootList:   ppRootList,
	}
	var resp netrDfsRemove2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsRemove2: %w", err)
		return
	}
	PpRootList = resp.PpRootList
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsRemove2 failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
