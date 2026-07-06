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

// netrFileEnumRequest is the [in]/[in,out] parameter set of NetrFileEnum: the [unique]
// server name, the [unique] base-path and user-name filters, the [in,out] enumeration
// container (whose Level selects the info arm), the byte budget, and the optional
// [in,out,unique] resume handle.
type netrFileEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	BasePath              *ndr.WSTR `ndr:"unique"`
	UserName              *ndr.WSTR `ndr:"unique"`
	InfoStruct            mssrvs.FILE_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrFileEnumRequest) Opnum() uint16 { return srvsvc.OpnumNetrFileEnum }

// netrFileEnumResponse is the reply: the updated [in,out] container, the [out] total
// entry count, the updated [in,out,unique] resume handle, and the NET_API_STATUS return
// value.
type netrFileEnumResponse struct {
	InfoStruct   mssrvs.FILE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrFileEnum calls NetrFileEnum (opnum 9), enumerating the open files/resources on the
// server, optionally filtered by base path and user name ([MS-SRVS] 3.1.4.2). The
// enumeration is stateful: pass the returned resume handle back on the next call to
// continue, starting from 0. ERROR_MORE_DATA indicates more pages remain and is not
// treated as an error; the returned container, total, and resume handle are valid then.
func NetrFileEnum(rpc ndr.Invoker, serverName, basePath, userName string, info mssrvs.FILE_ENUM_STRUCT, preferedMaximumLength, resumeHandle uint32) (mssrvs.FILE_ENUM_STRUCT, uint32, uint32, error) {
	resume := ndr.DWORD(resumeHandle)
	req := &netrFileEnumRequest{
		ServerName:            optWStr(serverName),
		BasePath:              optWStr(basePath),
		UserName:              optWStr(userName),
		InfoStruct:            info,
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
		ResumeHandle:          &resume,
	}
	var resp netrFileEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.FILE_ENUM_STRUCT{}, 0, resumeHandle, fmt.Errorf("NetrFileEnum: %w", err)
	}
	var outResume uint32
	if resp.ResumeHandle != nil {
		outResume = uint32(*resp.ResumeHandle)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, uint32(resp.TotalEntries), outResume, fmt.Errorf("NetrFileEnum failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, uint32(resp.TotalEntries), outResume, nil
}
