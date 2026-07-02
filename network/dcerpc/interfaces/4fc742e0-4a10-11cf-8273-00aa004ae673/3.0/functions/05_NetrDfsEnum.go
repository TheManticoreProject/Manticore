package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdfsnm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dfsnm"
)

// netrDfsEnumRequest carries the [in] parameters of NetrDfsEnum.
type netrDfsEnumRequest struct {
	Level        ndr.DWORD
	PrefMaxLen   ndr.DWORD
	DfsEnum      *msdfsnm.DFS_INFO_ENUM_STRUCT `ndr:"unique"`
	ResumeHandle *ndr.DWORD                    `ndr:"unique"`
}

func (*netrDfsEnumRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsEnum }

// netrDfsEnumResponse carries the [out] parameters and return value of NetrDfsEnum.
type netrDfsEnumResponse struct {
	DfsEnum      *msdfsnm.DFS_INFO_ENUM_STRUCT `ndr:"unique"`
	ResumeHandle *ndr.DWORD                    `ndr:"unique"`
	Status       ndr.DWORD                     `ndr:"retval"`
}

// NetrDfsEnum calls NetrDfsEnum (opnum 5) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsEnum(rpc ndr.Invoker, level ndr.DWORD, prefMaxLen ndr.DWORD, dfsEnum *msdfsnm.DFS_INFO_ENUM_STRUCT, resumeHandle *ndr.DWORD) (DfsEnum *msdfsnm.DFS_INFO_ENUM_STRUCT, ResumeHandle *ndr.DWORD, err error) {
	req := &netrDfsEnumRequest{
		Level:        level,
		PrefMaxLen:   prefMaxLen,
		DfsEnum:      dfsEnum,
		ResumeHandle: resumeHandle,
	}
	var resp netrDfsEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsEnum: %w", err)
		return
	}
	DfsEnum = resp.DfsEnum
	ResumeHandle = resp.ResumeHandle
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsEnum failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
