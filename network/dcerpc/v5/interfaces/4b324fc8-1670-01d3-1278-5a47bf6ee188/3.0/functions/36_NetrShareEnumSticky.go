package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
)

// netrShareEnumStickyRequest is the [in] parameter set of NetrShareEnumSticky: the optional
// server name, the inline [in,out] SHARE_ENUM_STRUCT, the preferred maximum reply length,
// and the optional [in,out,unique] ResumeHandle pointer.
type netrShareEnumStickyRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	InfoStruct            structures.SHARE_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrShareEnumStickyRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareEnumSticky
}

// netrShareEnumStickyResponse is the reply: the [in,out] SHARE_ENUM_STRUCT, the total
// entry count, the [in,out,unique] ResumeHandle pointer, and the NET_API_STATUS return
// value.
type netrShareEnumStickyResponse struct {
	InfoStruct   structures.SHARE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrShareEnumSticky calls NetrShareEnumSticky (opnum 36), enumerating the sticky shares
// on the server ([MS-SRVS] 3.1.4.15).
func NetrShareEnumSticky(rpc *client.Client, serverName string, infoStruct structures.SHARE_ENUM_STRUCT, preferedMaximumLength ndr.DWORD, resumeHandle *ndr.DWORD) (structures.SHARE_ENUM_STRUCT, ndr.DWORD, *ndr.DWORD, error) {
	req := &netrShareEnumStickyRequest{
		ServerName:            optWStr(serverName),
		InfoStruct:            infoStruct,
		PreferedMaximumLength: preferedMaximumLength,
		ResumeHandle:          resumeHandle,
	}
	var resp netrShareEnumStickyResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SHARE_ENUM_STRUCT{}, 0, nil, fmt.Errorf("NetrShareEnumSticky: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, resp.TotalEntries, resp.ResumeHandle, fmt.Errorf("NetrShareEnumSticky failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.InfoStruct, resp.TotalEntries, resp.ResumeHandle, nil
}
