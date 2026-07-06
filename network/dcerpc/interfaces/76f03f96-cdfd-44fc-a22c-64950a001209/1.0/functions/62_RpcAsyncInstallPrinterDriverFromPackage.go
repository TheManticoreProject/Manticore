package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncInstallPrinterDriverFromPackageRequest carries the [in] parameters of RpcAsyncInstallPrinterDriverFromPackage.
type rpcAsyncInstallPrinterDriverFromPackageRequest struct {
	PszServer      *ndr.WSTR `ndr:"unique"`
	PszInfPath     *ndr.WSTR `ndr:"unique"`
	PszDriverName  ndr.WSTR
	PszEnvironment ndr.WSTR
	DwFlags        ndr.DWORD
}

func (*rpcAsyncInstallPrinterDriverFromPackageRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncInstallPrinterDriverFromPackage
}

// rpcAsyncInstallPrinterDriverFromPackageResponse carries the [out] parameters and return value of RpcAsyncInstallPrinterDriverFromPackage.
type rpcAsyncInstallPrinterDriverFromPackageResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncInstallPrinterDriverFromPackage calls RpcAsyncInstallPrinterDriverFromPackage (opnum 62) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncInstallPrinterDriverFromPackage(rpc ndr.Invoker, pszServer *ndr.WSTR, pszInfPath *ndr.WSTR, pszDriverName ndr.WSTR, pszEnvironment ndr.WSTR, dwFlags ndr.DWORD) (err error) {
	req := &rpcAsyncInstallPrinterDriverFromPackageRequest{
		PszServer:      pszServer,
		PszInfPath:     pszInfPath,
		PszDriverName:  pszDriverName,
		PszEnvironment: pszEnvironment,
		DwFlags:        dwFlags,
	}
	var resp rpcAsyncInstallPrinterDriverFromPackageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncInstallPrinterDriverFromPackage: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncInstallPrinterDriverFromPackage failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
