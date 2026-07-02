package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiNodeResourceTypeControlRequest carries the [in] parameters of ApiNodeResourceTypeControl.
type apiNodeResourceTypeControlRequest struct {
	HCluster             mscmrp.HCLUSTER_RPC
	LpszResourceTypeName ndr.WSTR
	HNode                mscmrp.HNODE_RPC
	DwControlCode        ndr.DWORD
	LpInBuffer           []uint8 `ndr:"ref,size_is=NInBufferSize"`
	NInBufferSize        ndr.DWORD
	NOutBufferSize       ndr.DWORD
}

func (*apiNodeResourceTypeControlRequest) Opnum() uint16 {
	return clusapi.OpnumApiNodeResourceTypeControl
}

// apiNodeResourceTypeControlResponse carries the [out] parameters and return value of ApiNodeResourceTypeControl.
type apiNodeResourceTypeControlResponse struct {
	LpOutBuffer     []uint8 `ndr:"ref,size_is=NOutBufferSize,varying"`
	LpBytesReturned ndr.DWORD
	LpcbRequired    ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiNodeResourceTypeControl calls ApiNodeResourceTypeControl (opnum 74) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiNodeResourceTypeControl(rpc ndr.Invoker, hCluster mscmrp.HCLUSTER_RPC, lpszResourceTypeName ndr.WSTR, hNode mscmrp.HNODE_RPC, dwControlCode ndr.DWORD, lpInBuffer []uint8, nInBufferSize ndr.DWORD, nOutBufferSize ndr.DWORD) (LpOutBuffer []uint8, LpBytesReturned ndr.DWORD, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiNodeResourceTypeControlRequest{
		HCluster:             hCluster,
		LpszResourceTypeName: lpszResourceTypeName,
		HNode:                hNode,
		DwControlCode:        dwControlCode,
		LpInBuffer:           lpInBuffer,
		NInBufferSize:        nInBufferSize,
		NOutBufferSize:       nOutBufferSize,
	}
	var resp apiNodeResourceTypeControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiNodeResourceTypeControl: %w", err)
		return
	}
	LpOutBuffer = resp.LpOutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiNodeResourceTypeControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
