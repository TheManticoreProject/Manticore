package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcGetPublisherResourceMetadataRequest carries the [in] parameters of EvtRpcGetPublisherResourceMetadata.
type evtRpcGetPublisherResourceMetadataRequest struct {
	Handle     mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA
	PropertyId ndr.DWORD
	Flags      ndr.DWORD
}

func (*evtRpcGetPublisherResourceMetadataRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetPublisherResourceMetadata
}

// evtRpcGetPublisherResourceMetadataResponse carries the [out] parameters and return value of EvtRpcGetPublisherResourceMetadata.
type evtRpcGetPublisherResourceMetadataResponse struct {
	PubMetadataProps mseven6.EvtRpcVariantList
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcGetPublisherResourceMetadata calls EvtRpcGetPublisherResourceMetadata (opnum 25) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetPublisherResourceMetadata(rpc ndr.Invoker, handle mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA, propertyId ndr.DWORD, flags ndr.DWORD) (PubMetadataProps mseven6.EvtRpcVariantList, err error) {
	req := &evtRpcGetPublisherResourceMetadataRequest{
		Handle:     handle,
		PropertyId: propertyId,
		Flags:      flags,
	}
	var resp evtRpcGetPublisherResourceMetadataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetPublisherResourceMetadata: %w", err)
		return
	}
	PubMetadataProps = resp.PubMetadataProps
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetPublisherResourceMetadata failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
