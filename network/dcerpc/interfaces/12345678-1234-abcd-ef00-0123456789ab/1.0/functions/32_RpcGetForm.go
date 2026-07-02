package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetFormRequest carries the [in] parameters of RpcGetForm.
type rpcGetFormRequest struct {
	HPrinter  structures.PRINTER_HANDLE
	PFormName ndr.WSTR
	Level     ndr.DWORD
	PForm     []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf     ndr.DWORD
}

func (*rpcGetFormRequest) Opnum() uint16 { return winspool.OpnumRpcGetForm }

// rpcGetFormResponse carries the [out] parameters and return value of RpcGetForm.
type rpcGetFormResponse struct {
	PForm     []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetForm calls RpcGetForm (opnum 32) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetForm(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pFormName ndr.WSTR, level ndr.DWORD, pForm []uint8, cbBuf ndr.DWORD) (PForm []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetFormRequest{
		HPrinter:  hPrinter,
		PFormName: pFormName,
		Level:     level,
		PForm:     pForm,
		CbBuf:     cbBuf,
	}
	var resp rpcGetFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetForm: %w", err)
		return
	}
	PForm = resp.PForm
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetForm failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
