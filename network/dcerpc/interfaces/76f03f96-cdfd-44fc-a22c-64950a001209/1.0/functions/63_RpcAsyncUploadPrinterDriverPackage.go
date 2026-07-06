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

// rpcAsyncUploadPrinterDriverPackageRequest carries the [in] parameters of RpcAsyncUploadPrinterDriverPackage.
type rpcAsyncUploadPrinterDriverPackageRequest struct {
	PszServer       *ndr.WSTR `ndr:"unique"`
	PszInfPath      ndr.WSTR
	PszEnvironment  ndr.WSTR
	DwFlags         ndr.DWORD
	PszDestInfPath  []uint16 `ndr:"ref,conformant"`
	PcchDestInfPath ndr.DWORD
}

func (*rpcAsyncUploadPrinterDriverPackageRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncUploadPrinterDriverPackage
}

// rpcAsyncUploadPrinterDriverPackageResponse carries the [out] parameters and return value of RpcAsyncUploadPrinterDriverPackage.
type rpcAsyncUploadPrinterDriverPackageResponse struct {
	PszDestInfPath  []uint16 `ndr:"ref,conformant"`
	PcchDestInfPath ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcAsyncUploadPrinterDriverPackage calls RpcAsyncUploadPrinterDriverPackage (opnum 63) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncUploadPrinterDriverPackage(rpc ndr.Invoker, pszServer *ndr.WSTR, pszInfPath ndr.WSTR, pszEnvironment ndr.WSTR, dwFlags ndr.DWORD, pszDestInfPath []uint16, pcchDestInfPath ndr.DWORD) (PszDestInfPath []uint16, PcchDestInfPath ndr.DWORD, err error) {
	req := &rpcAsyncUploadPrinterDriverPackageRequest{
		PszServer:       pszServer,
		PszInfPath:      pszInfPath,
		PszEnvironment:  pszEnvironment,
		DwFlags:         dwFlags,
		PszDestInfPath:  pszDestInfPath,
		PcchDestInfPath: pcchDestInfPath,
	}
	var resp rpcAsyncUploadPrinterDriverPackageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncUploadPrinterDriverPackage: %w", err)
		return
	}
	PszDestInfPath = resp.PszDestInfPath
	PcchDestInfPath = resp.PcchDestInfPath
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncUploadPrinterDriverPackage failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
