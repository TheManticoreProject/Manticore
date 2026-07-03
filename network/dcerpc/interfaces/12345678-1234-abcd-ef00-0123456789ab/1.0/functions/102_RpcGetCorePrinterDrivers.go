package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcGetCorePrinterDriversRequest carries the [in] parameters of RpcGetCorePrinterDrivers.
type rpcGetCorePrinterDriversRequest struct {
	PszServer                  *ndr.WSTR `ndr:"unique"`
	PszEnvironment             ndr.WSTR
	CchCoreDrivers             ndr.DWORD
	PszzCoreDriverDependencies []uint16 `ndr:"ref,size_is=CchCoreDrivers"`
	CCorePrinterDrivers        ndr.DWORD
}

func (*rpcGetCorePrinterDriversRequest) Opnum() uint16 { return winspool.OpnumRpcGetCorePrinterDrivers }

// rpcGetCorePrinterDriversResponse carries the [out] parameters and return value of RpcGetCorePrinterDrivers.
type rpcGetCorePrinterDriversResponse struct {
	PCorePrinterDrivers []msrprn.CORE_PRINTER_DRIVER `ndr:"ref,size_is=CCorePrinterDrivers"`
	Status              ndr.DWORD                    `ndr:"retval"`
}

// RpcGetCorePrinterDrivers calls RpcGetCorePrinterDrivers (opnum 102) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetCorePrinterDrivers(rpc ndr.Invoker, pszServer *ndr.WSTR, pszEnvironment ndr.WSTR, cchCoreDrivers ndr.DWORD, pszzCoreDriverDependencies []uint16, cCorePrinterDrivers ndr.DWORD) (PCorePrinterDrivers []msrprn.CORE_PRINTER_DRIVER, err error) {
	req := &rpcGetCorePrinterDriversRequest{
		PszServer:                  pszServer,
		PszEnvironment:             pszEnvironment,
		CchCoreDrivers:             cchCoreDrivers,
		PszzCoreDriverDependencies: pszzCoreDriverDependencies,
		CCorePrinterDrivers:        cCorePrinterDrivers,
	}
	var resp rpcGetCorePrinterDriversResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetCorePrinterDrivers: %w", err)
		return
	}
	PCorePrinterDrivers = resp.PCorePrinterDrivers
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetCorePrinterDrivers failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
