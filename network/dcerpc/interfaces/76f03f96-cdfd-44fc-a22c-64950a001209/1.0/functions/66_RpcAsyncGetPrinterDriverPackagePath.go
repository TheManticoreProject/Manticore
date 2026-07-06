package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncGetPrinterDriverPackagePathRequest carries the [in] parameters of RpcAsyncGetPrinterDriverPackagePath.
type rpcAsyncGetPrinterDriverPackagePathRequest struct {
	PszServer           *ndr.WSTR `ndr:"unique"`
	PszEnvironment      ndr.WSTR
	PszLanguage         *ndr.WSTR `ndr:"unique"`
	PszPackageID        ndr.WSTR
	PszDriverPackageCab []uint16 `ndr:"ref,size_is=CchDriverPackageCab"`
	CchDriverPackageCab ndr.DWORD
}

func (*rpcAsyncGetPrinterDriverPackagePathRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrinterDriverPackagePath
}

// rpcAsyncGetPrinterDriverPackagePathResponse carries the [out] parameters and return value of RpcAsyncGetPrinterDriverPackagePath.
type rpcAsyncGetPrinterDriverPackagePathResponse struct {
	PszDriverPackageCab []uint16 `ndr:"ref,size_is=CchDriverPackageCab"`
	PcchRequiredSize    ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinterDriverPackagePath calls RpcAsyncGetPrinterDriverPackagePath (opnum 66) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinterDriverPackagePath(rpc ndr.Invoker, pszServer *ndr.WSTR, pszEnvironment ndr.WSTR, pszLanguage *ndr.WSTR, pszPackageID ndr.WSTR, pszDriverPackageCab []uint16, cchDriverPackageCab ndr.DWORD) (PszDriverPackageCab []uint16, PcchRequiredSize ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterDriverPackagePathRequest{
		PszServer:           pszServer,
		PszEnvironment:      pszEnvironment,
		PszLanguage:         pszLanguage,
		PszPackageID:        pszPackageID,
		PszDriverPackageCab: pszDriverPackageCab,
		CchDriverPackageCab: cchDriverPackageCab,
	}
	var resp rpcAsyncGetPrinterDriverPackagePathResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinterDriverPackagePath: %w", err)
		return
	}
	PszDriverPackageCab = resp.PszDriverPackageCab
	PcchRequiredSize = resp.PcchRequiredSize
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinterDriverPackagePath failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
