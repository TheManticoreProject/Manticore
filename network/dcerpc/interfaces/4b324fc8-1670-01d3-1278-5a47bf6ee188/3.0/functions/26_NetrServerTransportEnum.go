package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerTransportEnumRequest is the [in]/[in,out] parameter set of
// NetrServerTransportEnum: the [unique] server name, the [in,out] transport-enumeration
// structure (whose Level selects the info arm), the byte budget, and the optional
// [in,out,unique] resume handle.
type netrServerTransportEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	InfoStruct            mssrvs.SERVER_XPORT_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrServerTransportEnumRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerTransportEnum }

// netrServerTransportEnumResponse is the reply: the updated [in,out] structure, the
// [out] total entry count, the updated [in,out,unique] resume handle, and the
// NET_API_STATUS return value.
type netrServerTransportEnumResponse struct {
	InfoStruct   mssrvs.SERVER_XPORT_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrServerTransportEnum calls NetrServerTransportEnum (opnum 26), enumerating the
// transport protocols the server is bound to ([MS-SRVS] 3.1.4.22). The enumeration is
// stateful: pass the returned resume handle back to continue, starting from 0.
// ERROR_MORE_DATA indicates more pages remain and is not treated as an error.
func NetrServerTransportEnum(rpc ndr.Invoker, serverName string, info mssrvs.SERVER_XPORT_ENUM_STRUCT, preferedMaximumLength, resumeHandle uint32) (mssrvs.SERVER_XPORT_ENUM_STRUCT, uint32, uint32, error) {
	resume := ndr.DWORD(resumeHandle)
	req := &netrServerTransportEnumRequest{
		ServerName:            optWStr(serverName),
		InfoStruct:            info,
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
		ResumeHandle:          &resume,
	}
	var resp netrServerTransportEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.SERVER_XPORT_ENUM_STRUCT{}, 0, resumeHandle, fmt.Errorf("NetrServerTransportEnum: %w", err)
	}
	var outResume uint32
	if resp.ResumeHandle != nil {
		outResume = uint32(*resp.ResumeHandle)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, uint32(resp.TotalEntries), outResume, fmt.Errorf("NetrServerTransportEnum failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, uint32(resp.TotalEntries), outResume, nil
}
