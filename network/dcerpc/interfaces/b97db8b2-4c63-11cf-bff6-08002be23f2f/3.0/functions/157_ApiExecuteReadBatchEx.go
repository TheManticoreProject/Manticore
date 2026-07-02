package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiExecuteReadBatchExRequest carries the [in] parameters of ApiExecuteReadBatchEx.
type apiExecuteReadBatchExRequest struct {
	HKey     mscmrp.HKEY_RPC
	CbInData ndr.DWORD
	LpInData []uint8 `ndr:"ref,size_is=CbInData"`
	Flags    ndr.DWORD
}

func (*apiExecuteReadBatchExRequest) Opnum() uint16 { return clusapi.OpnumApiExecuteReadBatchEx }

// apiExecuteReadBatchExResponse carries the [out] parameters and return value of ApiExecuteReadBatchEx.
type apiExecuteReadBatchExResponse struct {
	CbOutData  ndr.DWORD
	LpOutData  []uint8 `ndr:"ref,size_is=CbOutData"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiExecuteReadBatchEx calls ApiExecuteReadBatchEx (opnum 157) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiExecuteReadBatchEx(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, cbInData ndr.DWORD, lpInData []uint8, flags ndr.DWORD) (CbOutData ndr.DWORD, LpOutData []uint8, Rpc_status ndr.DWORD, err error) {
	req := &apiExecuteReadBatchExRequest{
		HKey:     hKey,
		CbInData: cbInData,
		LpInData: lpInData,
		Flags:    flags,
	}
	var resp apiExecuteReadBatchExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiExecuteReadBatchEx: %w", err)
		return
	}
	CbOutData = resp.CbOutData
	LpOutData = resp.LpOutData
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiExecuteReadBatchEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
