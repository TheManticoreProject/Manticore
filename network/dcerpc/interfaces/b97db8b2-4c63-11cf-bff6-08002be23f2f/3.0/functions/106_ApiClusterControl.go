package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiClusterControlRequest carries the [in] parameters of ApiClusterControl.
type apiClusterControlRequest struct {
	HCluster       mscmrp.HCLUSTER_RPC
	DwControlCode  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=NInBufferSize"`
	NInBufferSize  ndr.DWORD
	NOutBufferSize ndr.DWORD
}

func (*apiClusterControlRequest) Opnum() uint16 { return clusapi.OpnumApiClusterControl }

// apiClusterControlResponse carries the [out] parameters and return value of ApiClusterControl.
type apiClusterControlResponse struct {
	LpOutBuffer     []uint8 `ndr:"ref,size_is=NOutBufferSize,varying"`
	LpBytesReturned ndr.DWORD
	LpcbRequired    ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiClusterControl calls ApiClusterControl (opnum 106) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiClusterControl(rpc ndr.Invoker, hCluster mscmrp.HCLUSTER_RPC, dwControlCode ndr.DWORD, lpInBuffer []uint8, nInBufferSize ndr.DWORD, nOutBufferSize ndr.DWORD) (LpOutBuffer []uint8, LpBytesReturned ndr.DWORD, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiClusterControlRequest{
		HCluster:       hCluster,
		DwControlCode:  dwControlCode,
		LpInBuffer:     lpInBuffer,
		NInBufferSize:  nInBufferSize,
		NOutBufferSize: nOutBufferSize,
	}
	var resp apiClusterControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiClusterControl: %w", err)
		return
	}
	LpOutBuffer = resp.LpOutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiClusterControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
