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

// rpcAsyncDeletePrinterDriverPackageRequest carries the [in] parameters of RpcAsyncDeletePrinterDriverPackage.
type rpcAsyncDeletePrinterDriverPackageRequest struct {
	PszServer      *ndr.WSTR `ndr:"unique"`
	PszInfPath     ndr.WSTR
	PszEnvironment ndr.WSTR
}

func (*rpcAsyncDeletePrinterDriverPackageRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterDriverPackage
}

// rpcAsyncDeletePrinterDriverPackageResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterDriverPackage.
type rpcAsyncDeletePrinterDriverPackageResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterDriverPackage calls RpcAsyncDeletePrinterDriverPackage (opnum 67) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterDriverPackage(rpc ndr.Invoker, pszServer *ndr.WSTR, pszInfPath ndr.WSTR, pszEnvironment ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePrinterDriverPackageRequest{
		PszServer:      pszServer,
		PszInfPath:     pszInfPath,
		PszEnvironment: pszEnvironment,
	}
	var resp rpcAsyncDeletePrinterDriverPackageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriverPackage: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriverPackage failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
