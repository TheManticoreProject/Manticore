package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrShareEnumRequest is the [in] parameter set of NetrShareEnum: the optional server
// name, the inline [in,out] SHARE_ENUM_STRUCT, the preferred maximum reply length, and the
// optional [in,out,unique] ResumeHandle pointer.
type netrShareEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	InfoStruct            structures.SHARE_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrShareEnumRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareEnum
}

// netrShareEnumResponse is the reply: the [in,out] SHARE_ENUM_STRUCT, the total entry
// count, the [in,out,unique] ResumeHandle pointer, and the NET_API_STATUS return value.
type netrShareEnumResponse struct {
	InfoStruct   structures.SHARE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrShareEnum calls NetrShareEnum (opnum 15), enumerating shares on the server
// ([MS-SRVS] 3.1.4.8).
func NetrShareEnum(rpc ndr.Invoker, serverName string, infoStruct structures.SHARE_ENUM_STRUCT, preferedMaximumLength ndr.DWORD, resumeHandle *ndr.DWORD) (structures.SHARE_ENUM_STRUCT, ndr.DWORD, *ndr.DWORD, error) {
	req := &netrShareEnumRequest{
		ServerName:            optWStr(serverName),
		InfoStruct:            infoStruct,
		PreferedMaximumLength: preferedMaximumLength,
		ResumeHandle:          resumeHandle,
	}
	var resp netrShareEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SHARE_ENUM_STRUCT{}, 0, nil, fmt.Errorf("NetrShareEnum: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, resp.TotalEntries, resp.ResumeHandle, fmt.Errorf("NetrShareEnum failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.InfoStruct, resp.TotalEntries, resp.ResumeHandle, nil
}
