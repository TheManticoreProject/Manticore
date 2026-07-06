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

// netrUseEnumRequest carries the [in] parameters of NetrUseEnum.
type netrUseEnumRequest struct {
	ServerName             *ndr.WSTR `ndr:"unique"`
	InfoStruct             mswkst.USE_ENUM_STRUCT
	PreferredMaximumLength ndr.DWORD
	ResumeHandle           *ndr.DWORD `ndr:"unique"`
}

func (*netrUseEnumRequest) Opnum() uint16 { return wkssvc.OpnumNetrUseEnum }

// netrUseEnumResponse carries the [out] parameters and return value of NetrUseEnum.
type netrUseEnumResponse struct {
	InfoStruct   mswkst.USE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrUseEnum calls NetrUseEnum (opnum 11) ([MS-WKST] 3.2.4).
func NetrUseEnum(rpc ndr.Invoker, serverName *ndr.WSTR, infoStruct mswkst.USE_ENUM_STRUCT, preferredMaximumLength ndr.DWORD, resumeHandle *ndr.DWORD) (InfoStruct mswkst.USE_ENUM_STRUCT, TotalEntries ndr.DWORD, ResumeHandle *ndr.DWORD, err error) {
	req := &netrUseEnumRequest{
		ServerName:             serverName,
		InfoStruct:             infoStruct,
		PreferredMaximumLength: preferredMaximumLength,
		ResumeHandle:           resumeHandle,
	}
	var resp netrUseEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUseEnum: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	TotalEntries = resp.TotalEntries
	ResumeHandle = resp.ResumeHandle
	// ERROR_MORE_DATA is a non-fatal status for the resume-capable Enum* methods: it
	// signals that the returned buffer holds a partial set and the caller should re-call
	// with the updated ResumeHandle. Treat it as success and surface the partial results.
	if s := uint32(resp.Status); s != wkssvc.StatusSuccess && s != wkssvc.ErrorMoreData {
		err = fmt.Errorf("NetrUseEnum failed: %s", wkssvc.StatusString(s))
	}
	return
}
