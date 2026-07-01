package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePrinterKeyRequest carries the [in] parameters of RpcDeletePrinterKey.
type rpcDeletePrinterKeyRequest struct {
	HPrinter structures.PRINTER_HANDLE
	PKeyName ndr.WSTR
}

func (*rpcDeletePrinterKeyRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterKey }

// rpcDeletePrinterKeyResponse carries the [out] parameters and return value of RpcDeletePrinterKey.
type rpcDeletePrinterKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterKey calls RpcDeletePrinterKey (opnum 82) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterKey(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pKeyName ndr.WSTR) (err error) {
	req := &rpcDeletePrinterKeyRequest{
		HPrinter: hPrinter,
		PKeyName: pKeyName,
	}
	var resp rpcDeletePrinterKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterKey: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterKey failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
