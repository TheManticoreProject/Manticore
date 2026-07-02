package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOnlineGroupExRequest carries the [in] parameters of ApiOnlineGroupEx.
type apiOnlineGroupExRequest struct {
	HGroup         mscmrp.HGROUP_RPC
	DwOnlineFlags  ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=CbInBufferSize"`
	CbInBufferSize ndr.DWORD
}

func (*apiOnlineGroupExRequest) Opnum() uint16 { return clusapi.OpnumApiOnlineGroupEx }

// apiOnlineGroupExResponse carries the [out] parameters and return value of ApiOnlineGroupEx.
type apiOnlineGroupExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOnlineGroupEx calls ApiOnlineGroupEx (opnum 130) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOnlineGroupEx(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, dwOnlineFlags ndr.DWORD, lpInBuffer []uint8, cbInBufferSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiOnlineGroupExRequest{
		HGroup:         hGroup,
		DwOnlineFlags:  dwOnlineFlags,
		LpInBuffer:     lpInBuffer,
		CbInBufferSize: cbInBufferSize,
	}
	var resp apiOnlineGroupExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOnlineGroupEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOnlineGroupEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
