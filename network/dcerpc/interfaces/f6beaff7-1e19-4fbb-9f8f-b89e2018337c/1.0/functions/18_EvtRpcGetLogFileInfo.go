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

// evtRpcGetLogFileInfoRequest carries the [in] parameters of EvtRpcGetLogFileInfo.
type evtRpcGetLogFileInfoRequest struct {
	LogHandle               mseven6.PCONTEXT_HANDLE_LOG_HANDLE
	PropertyId              ndr.DWORD
	PropertyValueBufferSize ndr.DWORD
}

func (*evtRpcGetLogFileInfoRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcGetLogFileInfo }

// evtRpcGetLogFileInfoResponse carries the [out] parameters and return value of EvtRpcGetLogFileInfo.
type evtRpcGetLogFileInfoResponse struct {
	PropertyValueBuffer       []uint8 `ndr:"ref,size_is=PropertyValueBufferSize"`
	PropertyValueBufferLength ndr.DWORD
	Status                    ndr.DWORD `ndr:"retval"`
}

// EvtRpcGetLogFileInfo calls EvtRpcGetLogFileInfo (opnum 18) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetLogFileInfo(rpc ndr.Invoker, logHandle mseven6.PCONTEXT_HANDLE_LOG_HANDLE, propertyId ndr.DWORD, propertyValueBufferSize ndr.DWORD) (PropertyValueBuffer []uint8, PropertyValueBufferLength ndr.DWORD, err error) {
	req := &evtRpcGetLogFileInfoRequest{
		LogHandle:               logHandle,
		PropertyId:              propertyId,
		PropertyValueBufferSize: propertyValueBufferSize,
	}
	var resp evtRpcGetLogFileInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetLogFileInfo: %w", err)
		return
	}
	PropertyValueBuffer = resp.PropertyValueBuffer
	PropertyValueBufferLength = resp.PropertyValueBufferLength
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetLogFileInfo failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
