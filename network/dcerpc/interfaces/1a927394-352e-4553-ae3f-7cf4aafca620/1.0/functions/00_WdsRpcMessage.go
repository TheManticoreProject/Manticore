package functions

import (
	"fmt"

	WdsRpcInterface "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a927394-352e-4553-ae3f-7cf4aafca620/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// wdsRpcMessageRequest carries the [in] parameters of WdsRpcMessage.
type wdsRpcMessageRequest struct {
	URequestPacketSize ndr.DWORD
	BRequestPacket     []uint8 `ndr:"ref,size_is=URequestPacketSize"`
}

func (*wdsRpcMessageRequest) Opnum() uint16 { return WdsRpcInterface.OpnumWdsRpcMessage }

// wdsRpcMessageResponse carries the [out] parameters and return value of WdsRpcMessage.
//
// pbReplyPacket is IDL "[out, size_is(, *puReplyPacketSize)] byte** pbReplyPacket": the
// server returns a [unique] pointer to a conformant array of *puReplyPacketSize bytes, so
// it is modeled as a []uint8 sized by the sibling PuReplyPacketSize field (not []*uint8).
type wdsRpcMessageResponse struct {
	PuReplyPacketSize ndr.DWORD
	PbReplyPacket     []uint8   `ndr:"unique,size_is=PuReplyPacketSize"`
	Status            ndr.DWORD `ndr:"retval"`
}

// WdsRpcMessage calls WdsRpcMessage (opnum 0) ([MS-WDSC] 3.1.4.1). It sends a request
// packet (constructed per [MS-WDSC] 2.2.1) and returns the server's reply packet. The
// method returns ERROR_SUCCESS (0) on success or a non-zero Win32 error code on failure.
func WdsRpcMessage(rpc ndr.Invoker, uRequestPacketSize ndr.DWORD, bRequestPacket []uint8) (PuReplyPacketSize ndr.DWORD, PbReplyPacket []uint8, err error) {
	req := &wdsRpcMessageRequest{
		URequestPacketSize: uRequestPacketSize,
		BRequestPacket:     bRequestPacket,
	}
	var resp wdsRpcMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WdsRpcMessage: %w", err)
		return
	}
	PuReplyPacketSize = resp.PuReplyPacketSize
	PbReplyPacket = resp.PbReplyPacket
	if uint32(resp.Status) != WdsRpcInterface.StatusSuccess {
		err = fmt.Errorf("WdsRpcMessage failed: %s", WdsRpcInterface.StatusString(uint32(resp.Status)))
	}
	return
}
