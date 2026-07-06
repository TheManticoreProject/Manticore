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

// evtRpcMessageRenderRequest carries the [in] parameters of EvtRpcMessageRender.
type evtRpcMessageRenderRequest struct {
	PubCfgObj     mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA
	SizeEventId   ndr.DWORD
	EventId       []uint8 `ndr:"ref,size_is=SizeEventId"`
	MessageId     ndr.DWORD
	Values        mseven6.EvtRpcVariantList
	Flags         ndr.DWORD
	MaxSizeString ndr.DWORD
}

func (*evtRpcMessageRenderRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcMessageRender }

// evtRpcMessageRenderResponse carries the [out] parameters and return value of EvtRpcMessageRender.
type evtRpcMessageRenderResponse struct {
	ActualSizeString ndr.DWORD
	NeededSizeString ndr.DWORD
	String           []uint8 `ndr:"unique,size_is=ActualSizeString"`
	Error            mseven6.RpcInfo
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcMessageRender calls EvtRpcMessageRender (opnum 9) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcMessageRender(rpc ndr.Invoker, pubCfgObj mseven6.PCONTEXT_HANDLE_PUBLISHER_METADATA, sizeEventId ndr.DWORD, eventId []uint8, messageId ndr.DWORD, values mseven6.EvtRpcVariantList, flags ndr.DWORD, maxSizeString ndr.DWORD) (ActualSizeString ndr.DWORD, NeededSizeString ndr.DWORD, String []uint8, Error mseven6.RpcInfo, err error) {
	req := &evtRpcMessageRenderRequest{
		PubCfgObj:     pubCfgObj,
		SizeEventId:   sizeEventId,
		EventId:       eventId,
		MessageId:     messageId,
		Values:        values,
		Flags:         flags,
		MaxSizeString: maxSizeString,
	}
	var resp evtRpcMessageRenderResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcMessageRender: %w", err)
		return
	}
	ActualSizeString = resp.ActualSizeString
	NeededSizeString = resp.NeededSizeString
	String = resp.String
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcMessageRender failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
