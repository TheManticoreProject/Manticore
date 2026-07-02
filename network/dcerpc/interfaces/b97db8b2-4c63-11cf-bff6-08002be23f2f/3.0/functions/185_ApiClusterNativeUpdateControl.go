package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiClusterNativeUpdateControlRequest carries the [in] parameters of ApiClusterNativeUpdateControl.
type apiClusterNativeUpdateControlRequest struct {
	InBuffer      uint8
	InBufferSize  ndr.DWORD
	OutBufferSize ndr.DWORD
}

func (*apiClusterNativeUpdateControlRequest) Opnum() uint16 {
	return clusapi.OpnumApiClusterNativeUpdateControl
}

// apiClusterNativeUpdateControlResponse carries the [out] parameters and return value of ApiClusterNativeUpdateControl.
type apiClusterNativeUpdateControlResponse struct {
	OutBuffer       uint8
	LpBytesReturned ndr.DWORD
	LpBytesNeeded   ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiClusterNativeUpdateControl calls ApiClusterNativeUpdateControl (opnum 185) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiClusterNativeUpdateControl(rpc ndr.Invoker, inBuffer uint8, inBufferSize ndr.DWORD, outBufferSize ndr.DWORD) (OutBuffer uint8, LpBytesReturned ndr.DWORD, LpBytesNeeded ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiClusterNativeUpdateControlRequest{
		InBuffer:      inBuffer,
		InBufferSize:  inBufferSize,
		OutBufferSize: outBufferSize,
	}
	var resp apiClusterNativeUpdateControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiClusterNativeUpdateControl: %w", err)
		return
	}
	OutBuffer = resp.OutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpBytesNeeded = resp.LpBytesNeeded
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiClusterNativeUpdateControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
