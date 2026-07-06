package functions

// IDL source: [MS-EVEN6] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even6/2d808edd-719a-4c69-b34a-df766adb5f0c
// A fetched copy is kept at ms-even6.idl in the interface directory.

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcGetNextEventMetadataRequest carries the [in] parameters of EvtRpcGetNextEventMetadata.
type evtRpcGetNextEventMetadataRequest struct {
	EventMetaDataEnum mseven6.PCONTEXT_HANDLE_EVENT_METADATA_ENUM
	Flags             ndr.DWORD
	NumRequested      ndr.DWORD
}

func (*evtRpcGetNextEventMetadataRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetNextEventMetadata
}

// evtRpcGetNextEventMetadataResponse carries the [out] parameters and return value of EvtRpcGetNextEventMetadata.
type evtRpcGetNextEventMetadataResponse struct {
	NumReturned            ndr.DWORD
	EventMetadataInstances []mseven6.EvtRpcVariantList `ndr:"unique,size_is=NumReturned"`
	Status                 ndr.DWORD                   `ndr:"retval"`
}

// EvtRpcGetNextEventMetadata calls EvtRpcGetNextEventMetadata (opnum 27) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetNextEventMetadata(rpc ndr.Invoker, eventMetaDataEnum mseven6.PCONTEXT_HANDLE_EVENT_METADATA_ENUM, flags ndr.DWORD, numRequested ndr.DWORD) (NumReturned ndr.DWORD, EventMetadataInstances []mseven6.EvtRpcVariantList, err error) {
	req := &evtRpcGetNextEventMetadataRequest{
		EventMetaDataEnum: eventMetaDataEnum,
		Flags:             flags,
		NumRequested:      numRequested,
	}
	var resp evtRpcGetNextEventMetadataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetNextEventMetadata: %w", err)
		return
	}
	NumReturned = resp.NumReturned
	EventMetadataInstances = resp.EventMetadataInstances
	// Metadata enumeration reports an exhausted enumerator through the return code; it is
	// benign — the caller inspects NumReturned ([MS-EVEN6] 3.1.4.31).
	switch uint32(resp.Status) {
	case IEventService.StatusSuccess, IEventService.ErrorNoMoreItems:
	default:
		err = fmt.Errorf("EvtRpcGetNextEventMetadata failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
