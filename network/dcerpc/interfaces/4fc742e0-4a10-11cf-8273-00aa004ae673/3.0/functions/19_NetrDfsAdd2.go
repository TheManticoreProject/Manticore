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

// netrDfsAdd2Request carries the [in] parameters of NetrDfsAdd2.
type netrDfsAdd2Request struct {
	DfsEntryPath ndr.WSTR
	DcName       ndr.WSTR
	ServerName   ndr.WSTR
	ShareName    *ndr.WSTR `ndr:"unique"`
	Comment      *ndr.WSTR `ndr:"unique"`
	Flags        ndr.DWORD
	PpRootList   *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
}

func (*netrDfsAdd2Request) Opnum() uint16 { return netdfs.OpnumNetrDfsAdd2 }

// netrDfsAdd2Response carries the [out] parameters and return value of NetrDfsAdd2.
type netrDfsAdd2Response struct {
	PpRootList *msdfsnm.DFSM_ROOT_LIST `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// NetrDfsAdd2 calls NetrDfsAdd2 (opnum 19) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAdd2(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, dcName ndr.WSTR, serverName ndr.WSTR, shareName *ndr.WSTR, comment *ndr.WSTR, flags ndr.DWORD, ppRootList *msdfsnm.DFSM_ROOT_LIST) (PpRootList *msdfsnm.DFSM_ROOT_LIST, err error) {
	req := &netrDfsAdd2Request{
		DfsEntryPath: dfsEntryPath,
		DcName:       dcName,
		ServerName:   serverName,
		ShareName:    shareName,
		Comment:      comment,
		Flags:        flags,
		PpRootList:   ppRootList,
	}
	var resp netrDfsAdd2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAdd2: %w", err)
		return
	}
	PpRootList = resp.PpRootList
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAdd2 failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
