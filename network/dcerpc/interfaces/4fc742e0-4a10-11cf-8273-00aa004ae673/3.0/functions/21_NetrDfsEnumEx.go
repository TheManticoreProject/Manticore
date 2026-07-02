package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdfsnm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dfsnm"
)

// netrDfsEnumExRequest carries the [in] parameters of NetrDfsEnumEx.
type netrDfsEnumExRequest struct {
	DfsEntryPath ndr.WSTR
	Level        ndr.DWORD
	PrefMaxLen   ndr.DWORD
	DfsEnum      *msdfsnm.DFS_INFO_ENUM_STRUCT `ndr:"unique"`
	ResumeHandle *ndr.DWORD                    `ndr:"unique"`
}

func (*netrDfsEnumExRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsEnumEx }

// netrDfsEnumExResponse carries the [out] parameters and return value of NetrDfsEnumEx.
type netrDfsEnumExResponse struct {
	DfsEnum      *msdfsnm.DFS_INFO_ENUM_STRUCT `ndr:"unique"`
	ResumeHandle *ndr.DWORD                    `ndr:"unique"`
	Status       ndr.DWORD                     `ndr:"retval"`
}

// NetrDfsEnumEx calls NetrDfsEnumEx (opnum 21) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsEnumEx(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, level ndr.DWORD, prefMaxLen ndr.DWORD, dfsEnum *msdfsnm.DFS_INFO_ENUM_STRUCT, resumeHandle *ndr.DWORD) (DfsEnum *msdfsnm.DFS_INFO_ENUM_STRUCT, ResumeHandle *ndr.DWORD, err error) {
	req := &netrDfsEnumExRequest{
		DfsEntryPath: dfsEntryPath,
		Level:        level,
		PrefMaxLen:   prefMaxLen,
		DfsEnum:      dfsEnum,
		ResumeHandle: resumeHandle,
	}
	var resp netrDfsEnumExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsEnumEx: %w", err)
		return
	}
	DfsEnum = resp.DfsEnum
	ResumeHandle = resp.ResumeHandle
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsEnumEx failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
