package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWkstaTransportEnumRequest carries the [in] parameters of NetrWkstaTransportEnum.
type netrWkstaTransportEnumRequest struct {
	ServerName             *ndr.WSTR `ndr:"unique"`
	TransportInfo          mswkst.WKSTA_TRANSPORT_ENUM_STRUCT
	PreferredMaximumLength ndr.DWORD
	ResumeHandle           *ndr.DWORD `ndr:"unique"`
}

func (*netrWkstaTransportEnumRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaTransportEnum }

// netrWkstaTransportEnumResponse carries the [out] parameters and return value of NetrWkstaTransportEnum.
type netrWkstaTransportEnumResponse struct {
	TransportInfo mswkst.WKSTA_TRANSPORT_ENUM_STRUCT
	TotalEntries  ndr.DWORD
	ResumeHandle  *ndr.DWORD `ndr:"unique"`
	Status        ndr.DWORD  `ndr:"retval"`
}

// NetrWkstaTransportEnum calls NetrWkstaTransportEnum (opnum 5) ([MS-WKST] 3.2.4).
func NetrWkstaTransportEnum(rpc ndr.Invoker, serverName *ndr.WSTR, transportInfo mswkst.WKSTA_TRANSPORT_ENUM_STRUCT, preferredMaximumLength ndr.DWORD, resumeHandle *ndr.DWORD) (TransportInfo mswkst.WKSTA_TRANSPORT_ENUM_STRUCT, TotalEntries ndr.DWORD, ResumeHandle *ndr.DWORD, err error) {
	req := &netrWkstaTransportEnumRequest{
		ServerName:             serverName,
		TransportInfo:          transportInfo,
		PreferredMaximumLength: preferredMaximumLength,
		ResumeHandle:           resumeHandle,
	}
	var resp netrWkstaTransportEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaTransportEnum: %w", err)
		return
	}
	TransportInfo = resp.TransportInfo
	TotalEntries = resp.TotalEntries
	ResumeHandle = resp.ResumeHandle
	// ERROR_MORE_DATA is a non-fatal status for the resume-capable Enum* methods: it
	// signals that the returned buffer holds a partial set and the caller should re-call
	// with the updated ResumeHandle. Treat it as success and surface the partial results.
	if s := uint32(resp.Status); s != wkssvc.StatusSuccess && s != wkssvc.ErrorMoreData {
		err = fmt.Errorf("NetrWkstaTransportEnum failed: %s", wkssvc.StatusString(s))
	}
	return
}
