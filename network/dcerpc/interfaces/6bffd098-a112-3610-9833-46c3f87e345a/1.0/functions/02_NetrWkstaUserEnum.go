package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWkstaUserEnumRequest carries the [in] parameters of NetrWkstaUserEnum.
type netrWkstaUserEnumRequest struct {
	ServerName             *ndr.WSTR `ndr:"unique"`
	UserInfo               mswkst.WKSTA_USER_ENUM_STRUCT
	PreferredMaximumLength ndr.DWORD
	ResumeHandle           *ndr.DWORD `ndr:"unique"`
}

func (*netrWkstaUserEnumRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaUserEnum }

// netrWkstaUserEnumResponse carries the [out] parameters and return value of NetrWkstaUserEnum.
type netrWkstaUserEnumResponse struct {
	UserInfo     mswkst.WKSTA_USER_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrWkstaUserEnum calls NetrWkstaUserEnum (opnum 2) ([MS-WKST] 3.2.4).
func NetrWkstaUserEnum(rpc ndr.Invoker, serverName *ndr.WSTR, userInfo mswkst.WKSTA_USER_ENUM_STRUCT, preferredMaximumLength ndr.DWORD, resumeHandle *ndr.DWORD) (UserInfo mswkst.WKSTA_USER_ENUM_STRUCT, TotalEntries ndr.DWORD, ResumeHandle *ndr.DWORD, err error) {
	req := &netrWkstaUserEnumRequest{
		ServerName:             serverName,
		UserInfo:               userInfo,
		PreferredMaximumLength: preferredMaximumLength,
		ResumeHandle:           resumeHandle,
	}
	var resp netrWkstaUserEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaUserEnum: %w", err)
		return
	}
	UserInfo = resp.UserInfo
	TotalEntries = resp.TotalEntries
	ResumeHandle = resp.ResumeHandle
	// ERROR_MORE_DATA is a non-fatal status for the resume-capable Enum* methods: it
	// signals that the returned buffer holds a partial set and the caller should re-call
	// with the updated ResumeHandle. Treat it as success and surface the partial results.
	if s := uint32(resp.Status); s != wkssvc.StatusSuccess && s != wkssvc.ErrorMoreData {
		err = fmt.Errorf("NetrWkstaUserEnum failed: %s", wkssvc.StatusString(s))
	}
	return
}
