package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiExecuteReadBatchRequest carries the [in] parameters of ApiExecuteReadBatch.
type apiExecuteReadBatchRequest struct {
	HKey     mscmrp.HKEY_RPC
	CbInData ndr.DWORD
	LpInData []uint8 `ndr:"ref,size_is=CbInData"`
}

func (*apiExecuteReadBatchRequest) Opnum() uint16 { return clusapi.OpnumApiExecuteReadBatch }

// apiExecuteReadBatchResponse carries the [out] parameters and return value of ApiExecuteReadBatch.
type apiExecuteReadBatchResponse struct {
	CbOutData  ndr.DWORD
	LpOutData  []uint8 `ndr:"ref,size_is=CbOutData"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiExecuteReadBatch calls ApiExecuteReadBatch (opnum 145) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiExecuteReadBatch(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, cbInData ndr.DWORD, lpInData []uint8) (CbOutData ndr.DWORD, LpOutData []uint8, Rpc_status ndr.DWORD, err error) {
	req := &apiExecuteReadBatchRequest{
		HKey:     hKey,
		CbInData: cbInData,
		LpInData: lpInData,
	}
	var resp apiExecuteReadBatchResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiExecuteReadBatch: %w", err)
		return
	}
	CbOutData = resp.CbOutData
	LpOutData = resp.LpOutData
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiExecuteReadBatch failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
