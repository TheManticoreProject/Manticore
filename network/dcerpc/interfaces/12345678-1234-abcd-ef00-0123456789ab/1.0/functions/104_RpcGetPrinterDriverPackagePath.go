package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetPrinterDriverPackagePathRequest carries the [in] parameters of RpcGetPrinterDriverPackagePath.
type rpcGetPrinterDriverPackagePathRequest struct {
	PszServer           *ndr.WSTR `ndr:"unique"`
	PszEnvironment      ndr.WSTR
	PszLanguage         *ndr.WSTR `ndr:"unique"`
	PszPackageID        ndr.WSTR
	PszDriverPackageCab []uint16 `ndr:"unique,size_is=CchDriverPackageCab"`
	CchDriverPackageCab ndr.DWORD
}

func (*rpcGetPrinterDriverPackagePathRequest) Opnum() uint16 {
	return winspool.OpnumRpcGetPrinterDriverPackagePath
}

// rpcGetPrinterDriverPackagePathResponse carries the [out] parameters and return value of RpcGetPrinterDriverPackagePath.
type rpcGetPrinterDriverPackagePathResponse struct {
	PszDriverPackageCab []uint16 `ndr:"unique,size_is=CchDriverPackageCab"`
	PcchRequiredSize    ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterDriverPackagePath calls RpcGetPrinterDriverPackagePath (opnum 104) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterDriverPackagePath(rpc ndr.Invoker, pszServer *ndr.WSTR, pszEnvironment ndr.WSTR, pszLanguage *ndr.WSTR, pszPackageID ndr.WSTR, pszDriverPackageCab []uint16, cchDriverPackageCab ndr.DWORD) (PszDriverPackageCab []uint16, PcchRequiredSize ndr.DWORD, err error) {
	req := &rpcGetPrinterDriverPackagePathRequest{
		PszServer:           pszServer,
		PszEnvironment:      pszEnvironment,
		PszLanguage:         pszLanguage,
		PszPackageID:        pszPackageID,
		PszDriverPackageCab: pszDriverPackageCab,
		CchDriverPackageCab: cchDriverPackageCab,
	}
	var resp rpcGetPrinterDriverPackagePathResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterDriverPackagePath: %w", err)
		return
	}
	PszDriverPackageCab = resp.PszDriverPackageCab
	PcchRequiredSize = resp.PcchRequiredSize
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterDriverPackagePath failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
