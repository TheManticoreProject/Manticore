package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetFormRequest carries the [in] parameters of RpcAsyncGetForm.
type rpcAsyncGetFormRequest struct {
	HPrinter  mspar.PRINTER_HANDLE
	PFormName ndr.WSTR
	Level     ndr.DWORD
	PForm     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf     ndr.DWORD
}

func (*rpcAsyncGetFormRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncGetForm }

// rpcAsyncGetFormResponse carries the [out] parameters and return value of RpcAsyncGetForm.
type rpcAsyncGetFormResponse struct {
	PForm     []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetForm calls RpcAsyncGetForm (opnum 23) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetForm(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pFormName ndr.WSTR, level ndr.DWORD, pForm []uint8, cbBuf ndr.DWORD) (PForm []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetFormRequest{
		HPrinter:  hPrinter,
		PFormName: pFormName,
		Level:     level,
		PForm:     pForm,
		CbBuf:     cbBuf,
	}
	var resp rpcAsyncGetFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetForm: %w", err)
		return
	}
	PForm = resp.PForm
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetForm failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
