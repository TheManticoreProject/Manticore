package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrGetJoinInformationRequest carries the [in] parameters of NetrGetJoinInformation.
// NameBuffer is [in,out,string] wchar_t**: a double pointer, modeled as a [unique] pointer
// to the wide string (referent id + deferred string on the wire, not an inline value).
type netrGetJoinInformationRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NameBuffer *ndr.WSTR `ndr:"unique"`
}

func (*netrGetJoinInformationRequest) Opnum() uint16 { return wkssvc.OpnumNetrGetJoinInformation }

// netrGetJoinInformationResponse carries the [out] parameters and return value of NetrGetJoinInformation.
type netrGetJoinInformationResponse struct {
	NameBuffer *ndr.WSTR `ndr:"unique"`
	BufferType mswkst.NETSETUP_JOIN_STATUS
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrGetJoinInformation calls NetrGetJoinInformation (opnum 20) ([MS-WKST] 3.2.4.3).
// It returns the workstation's join name and the join status (NETSETUP_JOIN_STATUS).
func NetrGetJoinInformation(rpc ndr.Invoker, serverName *ndr.WSTR, nameBuffer *ndr.WSTR) (NameBuffer *ndr.WSTR, BufferType mswkst.NETSETUP_JOIN_STATUS, err error) {
	req := &netrGetJoinInformationRequest{
		ServerName: serverName,
		NameBuffer: nameBuffer,
	}
	var resp netrGetJoinInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrGetJoinInformation: %w", err)
		return
	}
	NameBuffer = resp.NameBuffer
	BufferType = resp.BufferType
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrGetJoinInformation failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
