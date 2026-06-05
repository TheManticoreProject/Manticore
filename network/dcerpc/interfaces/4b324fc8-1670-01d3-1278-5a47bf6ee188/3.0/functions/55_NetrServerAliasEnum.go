package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrServerAliasEnumRequest is the [in]/[in,out] parameter set of NetrServerAliasEnum:
// the [unique] server name, the [in,out] alias-enumeration structure (whose Level
// selects the info arm), the byte budget, and the optional [in,out,unique] resume
// handle.
type netrServerAliasEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	InfoStruct            structures.SERVER_ALIAS_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrServerAliasEnumRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerAliasEnum }

// netrServerAliasEnumResponse is the reply: the updated [in,out] structure, the [out]
// total entry count, the updated [in,out,unique] resume handle, and the NET_API_STATUS
// return value.
type netrServerAliasEnumResponse struct {
	InfoStruct   structures.SERVER_ALIAS_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrServerAliasEnum calls NetrServerAliasEnum (opnum 55), enumerating the alias names
// configured on the server ([MS-SRVS] 3.1.4.29). The enumeration is stateful: pass the
// returned resume handle back to continue, starting from 0. ERROR_MORE_DATA indicates
// more pages remain and is not treated as an error.
func NetrServerAliasEnum(rpc ndr.Invoker, serverName string, info structures.SERVER_ALIAS_ENUM_STRUCT, preferedMaximumLength, resumeHandle uint32) (structures.SERVER_ALIAS_ENUM_STRUCT, uint32, uint32, error) {
	resume := ndr.DWORD(resumeHandle)
	req := &netrServerAliasEnumRequest{
		ServerName:            optWStr(serverName),
		InfoStruct:            info,
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
		ResumeHandle:          &resume,
	}
	var resp netrServerAliasEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SERVER_ALIAS_ENUM_STRUCT{}, 0, resumeHandle, fmt.Errorf("NetrServerAliasEnum: %w", err)
	}
	var outResume uint32
	if resp.ResumeHandle != nil {
		outResume = uint32(*resp.ResumeHandle)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, uint32(resp.TotalEntries), outResume, fmt.Errorf("NetrServerAliasEnum failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, uint32(resp.TotalEntries), outResume, nil
}
