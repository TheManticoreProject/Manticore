package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdfsnm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dfsnm"
)

// netrDfsSetInfo2Request carries the [in] parameters of NetrDfsSetInfo2.
type netrDfsSetInfo2Request struct {
	DfsEntryPath ndr.WSTR
	DcName       ndr.WSTR
	ServerName   *ndr.WSTR `ndr:"unique"`
	ShareName    *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PDfsInfo     msdfsnm.DFS_INFO_STRUCT
	PpRootList   *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
}

func (*netrDfsSetInfo2Request) Opnum() uint16 { return netdfs.OpnumNetrDfsSetInfo2 }

// netrDfsSetInfo2Response carries the [out] parameters and return value of NetrDfsSetInfo2.
type netrDfsSetInfo2Response struct {
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// NetrDfsSetInfo2 calls NetrDfsSetInfo2 (opnum 22) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsSetInfo2(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, dcName ndr.WSTR, serverName *ndr.WSTR, shareName *ndr.WSTR, level ndr.DWORD, pDfsInfo msdfsnm.DFS_INFO_STRUCT, ppRootList *msdfsnm.DFSM_ROOT_LIST) (PpRootList *msdfsnm.DFSM_ROOT_LIST, err error) {
	req := &netrDfsSetInfo2Request{
		DfsEntryPath: dfsEntryPath,
		DcName:       dcName,
		ServerName:   serverName,
		ShareName:    shareName,
		Level:        level,
		PDfsInfo:     pDfsInfo,
		PpRootList:   ppRootList,
	}
	var resp netrDfsSetInfo2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsSetInfo2: %w", err)
		return
	}
	PpRootList = resp.PpRootList
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsSetInfo2 failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
