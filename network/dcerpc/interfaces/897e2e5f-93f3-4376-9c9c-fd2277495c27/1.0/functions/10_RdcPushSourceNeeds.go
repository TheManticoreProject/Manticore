package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcPushSourceNeedsRequest carries the [in] parameters of RdcPushSourceNeeds.
type rdcPushSourceNeedsRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	SourceNeeds   []msfrs2.FRS_RDC_SOURCE_NEED `ndr:"ref,size_is=NeedCount"`
	NeedCount     ndr.DWORD
}

func (*rdcPushSourceNeedsRequest) Opnum() uint16 { return FrsTransport.OpnumRdcPushSourceNeeds }

// rdcPushSourceNeedsResponse carries the [out] parameters and return value of RdcPushSourceNeeds.
type rdcPushSourceNeedsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RdcPushSourceNeeds calls RdcPushSourceNeeds (opnum 10) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcPushSourceNeeds(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT, sourceNeeds []msfrs2.FRS_RDC_SOURCE_NEED, needCount ndr.DWORD) (err error) {
	req := &rdcPushSourceNeedsRequest{
		ServerContext: serverContext,
		SourceNeeds:   sourceNeeds,
		NeedCount:     needCount,
	}
	var resp rdcPushSourceNeedsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcPushSourceNeeds: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcPushSourceNeeds failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
