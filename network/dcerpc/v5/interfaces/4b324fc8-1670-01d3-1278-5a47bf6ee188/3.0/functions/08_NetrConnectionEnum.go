package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
)

// netrConnectionEnumRequest is the [in]/[in,out] parameter set of NetrConnectionEnum:
// the [unique] server name and share/computer qualifier, the [in,out] enumeration
// container (whose Level selects the info arm), the byte budget, and the optional
// [in,out,unique] resume handle.
type netrConnectionEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	Qualifier             *ndr.WSTR `ndr:"unique"`
	InfoStruct            structures.CONNECT_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrConnectionEnumRequest) Opnum() uint16 { return srvsvc.OpnumNetrConnectionEnum }

// netrConnectionEnumResponse is the reply: the updated [in,out] container, the [out]
// total entry count, the updated [in,out,unique] resume handle, and the NET_API_STATUS
// return value.
type netrConnectionEnumResponse struct {
	InfoStruct   structures.CONNECT_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrConnectionEnum calls NetrConnectionEnum (opnum 8), enumerating the connections
// made to a share or by a computer ([MS-SRVS] 3.1.4.1). The enumeration is stateful:
// pass the returned resume handle back on the next call to continue, starting from 0.
// ERROR_MORE_DATA indicates more pages remain and is not treated as an error; the
// returned container, total, and resume handle are valid in that case.
func NetrConnectionEnum(rpc *client.Client, serverName, qualifier string, info structures.CONNECT_ENUM_STRUCT, preferedMaximumLength, resumeHandle uint32) (structures.CONNECT_ENUM_STRUCT, uint32, uint32, error) {
	resume := ndr.DWORD(resumeHandle)
	req := &netrConnectionEnumRequest{
		ServerName:            optWStr(serverName),
		Qualifier:             optWStr(qualifier),
		InfoStruct:            info,
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
		ResumeHandle:          &resume,
	}
	var resp netrConnectionEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.CONNECT_ENUM_STRUCT{}, 0, resumeHandle, fmt.Errorf("NetrConnectionEnum: %w", err)
	}
	var outResume uint32
	if resp.ResumeHandle != nil {
		outResume = uint32(*resp.ResumeHandle)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, uint32(resp.TotalEntries), outResume, fmt.Errorf("NetrConnectionEnum failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, uint32(resp.TotalEntries), outResume, nil
}
