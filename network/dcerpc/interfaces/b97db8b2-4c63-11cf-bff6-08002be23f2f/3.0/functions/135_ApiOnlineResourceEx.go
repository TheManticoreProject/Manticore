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

// apiOnlineResourceExRequest carries the [in] parameters of ApiOnlineResourceEx.
type apiOnlineResourceExRequest struct {
	HResource      mscmrp.HRES_RPC
	DwOnlineFlags  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=CbInBufferSize"`
	CbInBufferSize ndr.DWORD
}

func (*apiOnlineResourceExRequest) Opnum() uint16 { return clusapi.OpnumApiOnlineResourceEx }

// apiOnlineResourceExResponse carries the [out] parameters and return value of ApiOnlineResourceEx.
type apiOnlineResourceExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOnlineResourceEx calls ApiOnlineResourceEx (opnum 135) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOnlineResourceEx(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, dwOnlineFlags ndr.DWORD, lpInBuffer []uint8, cbInBufferSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiOnlineResourceExRequest{
		HResource:      hResource,
		DwOnlineFlags:  dwOnlineFlags,
		LpInBuffer:     lpInBuffer,
		CbInBufferSize: cbInBufferSize,
	}
	var resp apiOnlineResourceExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOnlineResourceEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOnlineResourceEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
