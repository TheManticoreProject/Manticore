package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcGetPublisherMetadataRequest carries the [in] parameters of EvtRpcGetPublisherMetadata.
type evtRpcGetPublisherMetadataRequest struct {
	PublisherId *ndr.WSTR `ndr:"unique"`
	LogFilePath *ndr.WSTR `ndr:"unique"`
	Locale      ndr.DWORD
	Flags       ndr.DWORD
}

func (*evtRpcGetPublisherMetadataRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetPublisherMetadata
}

// evtRpcGetPublisherMetadataResponse carries the [out] parameters and return value of EvtRpcGetPublisherMetadata.
type evtRpcGetPublisherMetadataResponse struct {
	PubMetadataProps mseven6.EvtRpcVariantList
	PubMetadata      mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcGetPublisherMetadata calls EvtRpcGetPublisherMetadata (opnum 24) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetPublisherMetadata(rpc ndr.Invoker, publisherId *ndr.WSTR, logFilePath *ndr.WSTR, locale ndr.DWORD, flags ndr.DWORD) (PubMetadataProps mseven6.EvtRpcVariantList, PubMetadata mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA, err error) {
	req := &evtRpcGetPublisherMetadataRequest{
		PublisherId: publisherId,
		LogFilePath: logFilePath,
		Locale:      locale,
		Flags:       flags,
	}
	var resp evtRpcGetPublisherMetadataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetPublisherMetadata: %w", err)
		return
	}
	PubMetadataProps = resp.PubMetadataProps
	PubMetadata = resp.PubMetadata
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetPublisherMetadata failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
