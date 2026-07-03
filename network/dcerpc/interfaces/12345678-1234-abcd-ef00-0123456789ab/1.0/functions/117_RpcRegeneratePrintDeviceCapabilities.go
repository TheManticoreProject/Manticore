package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcRegeneratePrintDeviceCapabilitiesRequest carries the [in] parameters of RpcRegeneratePrintDeviceCapabilities.
type rpcRegeneratePrintDeviceCapabilitiesRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
}

func (*rpcRegeneratePrintDeviceCapabilitiesRequest) Opnum() uint16 {
	return winspool.OpnumRpcRegeneratePrintDeviceCapabilities
}

// rpcRegeneratePrintDeviceCapabilitiesResponse carries the [out] parameters and return value of RpcRegeneratePrintDeviceCapabilities.
type rpcRegeneratePrintDeviceCapabilitiesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcRegeneratePrintDeviceCapabilities calls RpcRegeneratePrintDeviceCapabilities (opnum 117) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRegeneratePrintDeviceCapabilities(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE) (err error) {
	req := &rpcRegeneratePrintDeviceCapabilitiesRequest{
		HPrinter: hPrinter,
	}
	var resp rpcRegeneratePrintDeviceCapabilitiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRegeneratePrintDeviceCapabilities: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRegeneratePrintDeviceCapabilities failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
