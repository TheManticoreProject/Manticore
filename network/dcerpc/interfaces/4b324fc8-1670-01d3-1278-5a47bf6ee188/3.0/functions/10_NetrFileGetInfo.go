package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrFileGetInfoRequest is the [in] parameter set of NetrFileGetInfo: the [unique]
// server name, the file identifier to query, and the information level selecting which
// FILE_INFO union arm is returned.
type netrFileGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	FileId     ndr.DWORD
	Level      ndr.DWORD
}

func (*netrFileGetInfoRequest) Opnum() uint16 { return srvsvc.OpnumNetrFileGetInfo }

// netrFileGetInfoResponse is the reply: the [out, switch_is(Level)] FILE_INFO union (a
// single ref pointer in the IDL, modelled inline) and the NET_API_STATUS return value.
type netrFileGetInfoResponse struct {
	InfoStruct mssrvs.FILE_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrFileGetInfo calls NetrFileGetInfo (opnum 10), returning information about a single
// open file/resource identified by FileId at the requested level ([MS-SRVS] 3.1.4.3).
func NetrFileGetInfo(rpc ndr.Invoker, serverName string, fileId, level uint32) (mssrvs.FILE_INFO, error) {
	req := &netrFileGetInfoRequest{
		ServerName: optWStr(serverName),
		FileId:     ndr.DWORD(fileId),
		Level:      ndr.DWORD(level),
	}
	var resp netrFileGetInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.FILE_INFO{}, fmt.Errorf("NetrFileGetInfo: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, fmt.Errorf("NetrFileGetInfo failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, nil
}
