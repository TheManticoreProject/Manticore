package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumFormsRequest carries the [in] parameters of RpcEnumForms.
type rpcEnumFormsRequest struct {
	HPrinter structures.PRINTER_HANDLE
	Level    ndr.DWORD
	PForm    []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcEnumFormsRequest) Opnum() uint16 { return winspool.OpnumRpcEnumForms }

// rpcEnumFormsResponse carries the [out] parameters and return value of RpcEnumForms.
type rpcEnumFormsResponse struct {
	PForm      []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumForms calls RpcEnumForms (opnum 34) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumForms(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, level ndr.DWORD, pForm []uint8, cbBuf ndr.DWORD) (PForm []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumFormsRequest{
		HPrinter: hPrinter,
		Level:    level,
		PForm:    pForm,
		CbBuf:    cbBuf,
	}
	var resp rpcEnumFormsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumForms: %w", err)
		return
	}
	PForm = resp.PForm
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumForms failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
