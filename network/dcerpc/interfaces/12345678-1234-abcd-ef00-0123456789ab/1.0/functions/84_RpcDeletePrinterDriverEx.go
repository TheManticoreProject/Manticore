package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePrinterDriverExRequest carries the [in] parameters of RpcDeletePrinterDriverEx.
type rpcDeletePrinterDriverExRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment ndr.WSTR
	PDriverName  ndr.WSTR
	DwDeleteFlag ndr.DWORD
	DwVersionNum ndr.DWORD
}

func (*rpcDeletePrinterDriverExRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterDriverEx }

// rpcDeletePrinterDriverExResponse carries the [out] parameters and return value of RpcDeletePrinterDriverEx.
type rpcDeletePrinterDriverExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterDriverEx calls RpcDeletePrinterDriverEx (opnum 84) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterDriverEx(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pDriverName ndr.WSTR, dwDeleteFlag ndr.DWORD, dwVersionNum ndr.DWORD) (err error) {
	req := &rpcDeletePrinterDriverExRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		PDriverName:  pDriverName,
		DwDeleteFlag: dwDeleteFlag,
		DwVersionNum: dwVersionNum,
	}
	var resp rpcDeletePrinterDriverExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterDriverEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterDriverEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
