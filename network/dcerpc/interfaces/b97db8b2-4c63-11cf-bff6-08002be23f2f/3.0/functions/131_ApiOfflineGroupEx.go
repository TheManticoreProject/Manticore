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

// apiOfflineGroupExRequest carries the [in] parameters of ApiOfflineGroupEx.
type apiOfflineGroupExRequest struct {
	HGroup         mscmrp.HGROUP_RPC
	DwOfflineFlags ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=CbInBufferSize"`
	CbInBufferSize ndr.DWORD
}

func (*apiOfflineGroupExRequest) Opnum() uint16 { return clusapi.OpnumApiOfflineGroupEx }

// apiOfflineGroupExResponse carries the [out] parameters and return value of ApiOfflineGroupEx.
type apiOfflineGroupExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOfflineGroupEx calls ApiOfflineGroupEx (opnum 131) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOfflineGroupEx(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, dwOfflineFlags ndr.DWORD, lpInBuffer []uint8, cbInBufferSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiOfflineGroupExRequest{
		HGroup:         hGroup,
		DwOfflineFlags: dwOfflineFlags,
		LpInBuffer:     lpInBuffer,
		CbInBufferSize: cbInBufferSize,
	}
	var resp apiOfflineGroupExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOfflineGroupEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOfflineGroupEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
