package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGroupControlRequest carries the [in] parameters of ApiGroupControl.
type apiGroupControlRequest struct {
	HGroup         mscmrp.HGROUP_RPC
	DwControlCode  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=NInBufferSize"`
	NInBufferSize  ndr.DWORD
	NOutBufferSize ndr.DWORD
}

func (*apiGroupControlRequest) Opnum() uint16 { return clusapi.OpnumApiGroupControl }

// apiGroupControlResponse carries the [out] parameters and return value of ApiGroupControl.
type apiGroupControlResponse struct {
	LpOutBuffer     []uint8 `ndr:"ref,size_is=NOutBufferSize,varying"`
	LpBytesReturned ndr.DWORD
	LpcbRequired    ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiGroupControl calls ApiGroupControl (opnum 77) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGroupControl(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, dwControlCode ndr.DWORD, lpInBuffer []uint8, nInBufferSize ndr.DWORD, nOutBufferSize ndr.DWORD) (LpOutBuffer []uint8, LpBytesReturned ndr.DWORD, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGroupControlRequest{
		HGroup:         hGroup,
		DwControlCode:  dwControlCode,
		LpInBuffer:     lpInBuffer,
		NInBufferSize:  nInBufferSize,
		NOutBufferSize: nOutBufferSize,
	}
	var resp apiGroupControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGroupControl: %w", err)
		return
	}
	LpOutBuffer = resp.LpOutBuffer
	LpBytesReturned = resp.LpBytesReturned
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGroupControl failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
