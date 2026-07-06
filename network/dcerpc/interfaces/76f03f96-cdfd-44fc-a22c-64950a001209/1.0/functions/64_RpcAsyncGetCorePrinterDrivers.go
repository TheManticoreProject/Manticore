package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetCorePrinterDriversRequest carries the [in] parameters of RpcAsyncGetCorePrinterDrivers.
type rpcAsyncGetCorePrinterDriversRequest struct {
	PszServer                  *ndr.WSTR `ndr:"unique"`
	PszEnvironment             ndr.WSTR
	CchCoreDrivers             ndr.DWORD
	PszzCoreDriverDependencies []uint16 `ndr:"ref,size_is=CchCoreDrivers"`
	CCorePrinterDrivers        ndr.DWORD
}

func (*rpcAsyncGetCorePrinterDriversRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetCorePrinterDrivers
}

// rpcAsyncGetCorePrinterDriversResponse carries the [out] parameters and return value of RpcAsyncGetCorePrinterDrivers.
type rpcAsyncGetCorePrinterDriversResponse struct {
	PCorePrinterDrivers []mspar.CORE_PRINTER_DRIVER `ndr:"ref,size_is=CCorePrinterDrivers"`
	Status              ndr.DWORD                   `ndr:"retval"`
}

// RpcAsyncGetCorePrinterDrivers calls RpcAsyncGetCorePrinterDrivers (opnum 64) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetCorePrinterDrivers(rpc ndr.Invoker, pszServer *ndr.WSTR, pszEnvironment ndr.WSTR, cchCoreDrivers ndr.DWORD, pszzCoreDriverDependencies []uint16, cCorePrinterDrivers ndr.DWORD) (PCorePrinterDrivers []mspar.CORE_PRINTER_DRIVER, err error) {
	req := &rpcAsyncGetCorePrinterDriversRequest{
		PszServer:                  pszServer,
		PszEnvironment:             pszEnvironment,
		CchCoreDrivers:             cchCoreDrivers,
		PszzCoreDriverDependencies: pszzCoreDriverDependencies,
		CCorePrinterDrivers:        cCorePrinterDrivers,
	}
	var resp rpcAsyncGetCorePrinterDriversResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetCorePrinterDrivers: %w", err)
		return
	}
	PCorePrinterDrivers = resp.PCorePrinterDrivers
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetCorePrinterDrivers failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
