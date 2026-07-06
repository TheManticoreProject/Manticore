package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncCorePrinterDriverInstalledRequest carries the [in] parameters of RpcAsyncCorePrinterDriverInstalled.
type rpcAsyncCorePrinterDriverInstalledRequest struct {
	PszServer        *ndr.WSTR `ndr:"unique"`
	PszEnvironment   ndr.WSTR
	CoreDriverGUID   guid.GUID
	FtDriverDate     mspar.FILETIME
	DwlDriverVersion uint64
}

func (*rpcAsyncCorePrinterDriverInstalledRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncCorePrinterDriverInstalled
}

// rpcAsyncCorePrinterDriverInstalledResponse carries the [out] parameters and return value of RpcAsyncCorePrinterDriverInstalled.
type rpcAsyncCorePrinterDriverInstalledResponse struct {
	PbDriverInstalled int32
	Status            ndr.DWORD `ndr:"retval"`
}

// RpcAsyncCorePrinterDriverInstalled calls RpcAsyncCorePrinterDriverInstalled (opnum 65) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncCorePrinterDriverInstalled(rpc ndr.Invoker, pszServer *ndr.WSTR, pszEnvironment ndr.WSTR, coreDriverGUID guid.GUID, ftDriverDate mspar.FILETIME, dwlDriverVersion uint64) (PbDriverInstalled int32, err error) {
	req := &rpcAsyncCorePrinterDriverInstalledRequest{
		PszServer:        pszServer,
		PszEnvironment:   pszEnvironment,
		CoreDriverGUID:   coreDriverGUID,
		FtDriverDate:     ftDriverDate,
		DwlDriverVersion: dwlDriverVersion,
	}
	var resp rpcAsyncCorePrinterDriverInstalledResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncCorePrinterDriverInstalled: %w", err)
		return
	}
	PbDriverInstalled = resp.PbDriverInstalled
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncCorePrinterDriverInstalled failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
