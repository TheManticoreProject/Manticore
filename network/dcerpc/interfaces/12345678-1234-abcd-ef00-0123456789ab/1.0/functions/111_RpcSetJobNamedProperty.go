package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcSetJobNamedPropertyRequest carries the [in] parameters of RpcSetJobNamedProperty.
type rpcSetJobNamedPropertyRequest struct {
	HPrinter  msrprn.PRINTER_HANDLE
	JobId     ndr.DWORD
	PProperty msrprn.RPC_PrintNamedProperty
}

func (*rpcSetJobNamedPropertyRequest) Opnum() uint16 { return winspool.OpnumRpcSetJobNamedProperty }

// rpcSetJobNamedPropertyResponse carries the [out] parameters and return value of RpcSetJobNamedProperty.
type rpcSetJobNamedPropertyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetJobNamedProperty calls RpcSetJobNamedProperty (opnum 111) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetJobNamedProperty(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, pProperty msrprn.RPC_PrintNamedProperty) (err error) {
	req := &rpcSetJobNamedPropertyRequest{
		HPrinter:  hPrinter,
		JobId:     jobId,
		PProperty: pProperty,
	}
	var resp rpcSetJobNamedPropertyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetJobNamedProperty: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetJobNamedProperty failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
