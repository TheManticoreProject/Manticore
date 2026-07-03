package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcEnumPrinterKeyRequest carries the [in] parameters of RpcEnumPrinterKey.
type rpcEnumPrinterKeyRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	PKeyName ndr.WSTR
	CbSubkey ndr.DWORD
}

func (*rpcEnumPrinterKeyRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrinterKey }

// rpcEnumPrinterKeyResponse carries the [out] parameters and return value of RpcEnumPrinterKey.
type rpcEnumPrinterKeyResponse struct {
	PSubkey   []uint16 `ndr:"ref,conformant"`
	PcbSubkey ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrinterKey calls RpcEnumPrinterKey (opnum 80) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrinterKey(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pKeyName ndr.WSTR, cbSubkey ndr.DWORD) (PSubkey []uint16, PcbSubkey ndr.DWORD, err error) {
	req := &rpcEnumPrinterKeyRequest{
		HPrinter: hPrinter,
		PKeyName: pKeyName,
		CbSubkey: cbSubkey,
	}
	var resp rpcEnumPrinterKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrinterKey: %w", err)
		return
	}
	PSubkey = resp.PSubkey
	PcbSubkey = resp.PcbSubkey
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrinterKey failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
