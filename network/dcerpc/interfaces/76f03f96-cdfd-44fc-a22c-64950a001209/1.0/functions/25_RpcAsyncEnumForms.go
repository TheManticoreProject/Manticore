package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEnumFormsRequest carries the [in] parameters of RpcAsyncEnumForms.
type rpcAsyncEnumFormsRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	Level    ndr.DWORD
	PForm    []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncEnumFormsRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncEnumForms }

// rpcAsyncEnumFormsResponse carries the [out] parameters and return value of RpcAsyncEnumForms.
type rpcAsyncEnumFormsResponse struct {
	PForm      []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumForms calls RpcAsyncEnumForms (opnum 25) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumForms(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, level ndr.DWORD, pForm []uint8, cbBuf ndr.DWORD) (PForm []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumFormsRequest{
		HPrinter: hPrinter,
		Level:    level,
		PForm:    pForm,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncEnumFormsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumForms: %w", err)
		return
	}
	PForm = resp.PForm
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumForms failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
