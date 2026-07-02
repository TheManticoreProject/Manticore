package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiNodeNetInterfaceControlRequest carries the [in] parameters of ApiNodeNetInterfaceControl.
type apiNodeNetInterfaceControlRequest struct {
	HNetInterface  mscmrp.HNETINTERFACE_RPC
	HNode          mscmrp.HNODE_RPC
	DwControlCode  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=NInBufferSize"`
	NInBufferSize  ndr.DWORD
	NOutBufferSize ndr.DWORD
}

func (*apiNodeNetInterfaceControlRequest) Opnum() uint16 {
	return clusapi.OpnumApiNodeNetInterfaceControl
}

// apiNodeNetInterfaceControlResponse carries the [out] parameters and return value of ApiNodeNetInterfaceControl.
type apiNodeNetInterfaceControlResponse struct {
	LpOutBuffer     []uint8 `ndr:"ref,size_is=NOutBufferSize,varying"`
	LpBytesReturned ndr.DWORD
	LpcbRequired    ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiNodeNetInterfaceControl calls ApiNodeNetInterfaceControl (opnum 97) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiNodeNetInterfaceControl(rpc ndr.Invoker, hNetInterface mscmrp.HNETINTERFACE_RPC, hNode mscmrp.HNODE_RPC, dwControlCode ndr.DWORD, lpInBuffer []uint8, nInBufferSize ndr.DWORD, nOutBufferSize ndr.DWORD) (LpOutBuffer []uint8, LpBytesReturned ndr.DWORD, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiNodeNetInterfaceControlRequest{
		HNetInterface:  hNetInterface,
		HNode:          hNode,
		DwControlCode:  dwControlCode,
		LpInBuffer:     lpInBuffer,
		NInBufferSize:  nInBufferSize,
		NOutBufferSize: nOutBufferSize,
	}
	var resp apiNodeNetInterfaceControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiNodeNetInterfaceControl: %w", err)
		return
	}
	LpOutBuffer = resp.LpOutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiNodeNetInterfaceControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
