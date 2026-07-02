package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiNodeGroupSetControlRequest carries the [in] parameters of ApiNodeGroupSetControl.
type apiNodeGroupSetControlRequest struct {
	HGroupSet      mscmrp.HGROUPSET_RPC
	HNode          mscmrp.HNODE_RPC
	DwControlCode  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=NInBufferSize"`
	NInBufferSize  ndr.DWORD
	NOutBufferSize ndr.DWORD
}

func (*apiNodeGroupSetControlRequest) Opnum() uint16 { return clusapi.OpnumApiNodeGroupSetControl }

// apiNodeGroupSetControlResponse carries the [out] parameters and return value of ApiNodeGroupSetControl.
type apiNodeGroupSetControlResponse struct {
	LpOutBuffer     []uint8 `ndr:"ref,size_is=NOutBufferSize,varying"`
	LpBytesReturned ndr.DWORD
	LpcbRequired    ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiNodeGroupSetControl calls ApiNodeGroupSetControl (opnum 173) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiNodeGroupSetControl(rpc ndr.Invoker, hGroupSet mscmrp.HGROUPSET_RPC, hNode mscmrp.HNODE_RPC, dwControlCode ndr.DWORD, lpInBuffer []uint8, nInBufferSize ndr.DWORD, nOutBufferSize ndr.DWORD) (LpOutBuffer []uint8, LpBytesReturned ndr.DWORD, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiNodeGroupSetControlRequest{
		HGroupSet:      hGroupSet,
		HNode:          hNode,
		DwControlCode:  dwControlCode,
		LpInBuffer:     lpInBuffer,
		NInBufferSize:  nInBufferSize,
		NOutBufferSize: nOutBufferSize,
	}
	var resp apiNodeGroupSetControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiNodeGroupSetControl: %w", err)
		return
	}
	LpOutBuffer = resp.LpOutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiNodeGroupSetControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
