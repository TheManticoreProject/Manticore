package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOfflineResourceExRequest carries the [in] parameters of ApiOfflineResourceEx.
type apiOfflineResourceExRequest struct {
	HResource      mscmrp.HRES_RPC
	DwOfflineFlags ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=CbInBufferSize"`
	CbInBufferSize ndr.DWORD
}

func (*apiOfflineResourceExRequest) Opnum() uint16 { return clusapi.OpnumApiOfflineResourceEx }

// apiOfflineResourceExResponse carries the [out] parameters and return value of ApiOfflineResourceEx.
type apiOfflineResourceExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOfflineResourceEx calls ApiOfflineResourceEx (opnum 136) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOfflineResourceEx(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, dwOfflineFlags ndr.DWORD, lpInBuffer []uint8, cbInBufferSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiOfflineResourceExRequest{
		HResource:      hResource,
		DwOfflineFlags: dwOfflineFlags,
		LpInBuffer:     lpInBuffer,
		CbInBufferSize: cbInBufferSize,
	}
	var resp apiOfflineResourceExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOfflineResourceEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOfflineResourceEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
