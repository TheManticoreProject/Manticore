package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcAddPrinterDriverExRequest carries the [in] parameters of RpcAddPrinterDriverEx.
type rpcAddPrinterDriverExRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PDriverContainer msrprn.DRIVER_CONTAINER
	DwFileCopyFlags  ndr.DWORD
}

func (*rpcAddPrinterDriverExRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrinterDriverEx }

// rpcAddPrinterDriverExResponse carries the [out] parameters and return value of RpcAddPrinterDriverEx.
type rpcAddPrinterDriverExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPrinterDriverEx calls RpcAddPrinterDriverEx (opnum 89) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrinterDriverEx(rpc ndr.Invoker, pName *ndr.WSTR, pDriverContainer msrprn.DRIVER_CONTAINER, dwFileCopyFlags ndr.DWORD) (err error) {
	req := &rpcAddPrinterDriverExRequest{
		PName:            pName,
		PDriverContainer: pDriverContainer,
		DwFileCopyFlags:  dwFileCopyFlags,
	}
	var resp rpcAddPrinterDriverExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPrinterDriverEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPrinterDriverEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
