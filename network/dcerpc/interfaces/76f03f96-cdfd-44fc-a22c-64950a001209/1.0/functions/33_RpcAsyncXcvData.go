package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncXcvDataRequest carries the [in] parameters of RpcAsyncXcvData.
type rpcAsyncXcvDataRequest struct {
	HXcv         mspar.PRINTER_HANDLE
	PszDataName  ndr.WSTR
	PInputData   []uint8 `ndr:"ref,size_is=CbInputData"`
	CbInputData  ndr.DWORD
	CbOutputData ndr.DWORD
	PdwStatus    ndr.DWORD
}

func (*rpcAsyncXcvDataRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncXcvData }

// rpcAsyncXcvDataResponse carries the [out] parameters and return value of RpcAsyncXcvData.
type rpcAsyncXcvDataResponse struct {
	POutputData     []uint8 `ndr:"ref,size_is=CbOutputData"`
	PcbOutputNeeded ndr.DWORD
	PdwStatus       ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcAsyncXcvData calls RpcAsyncXcvData (opnum 33) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncXcvData(rpc ndr.Invoker, hXcv mspar.PRINTER_HANDLE, pszDataName ndr.WSTR, pInputData []uint8, cbInputData ndr.DWORD, cbOutputData ndr.DWORD, pdwStatus ndr.DWORD) (POutputData []uint8, PcbOutputNeeded ndr.DWORD, PdwStatus ndr.DWORD, err error) {
	req := &rpcAsyncXcvDataRequest{
		HXcv:         hXcv,
		PszDataName:  pszDataName,
		PInputData:   pInputData,
		CbInputData:  cbInputData,
		CbOutputData: cbOutputData,
		PdwStatus:    pdwStatus,
	}
	var resp rpcAsyncXcvDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncXcvData: %w", err)
		return
	}
	POutputData = resp.POutputData
	PcbOutputNeeded = resp.PcbOutputNeeded
	PdwStatus = resp.PdwStatus
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncXcvData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
