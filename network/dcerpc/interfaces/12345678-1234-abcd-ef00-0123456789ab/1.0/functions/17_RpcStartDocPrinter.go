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

// rpcStartDocPrinterRequest carries the [in] parameters of RpcStartDocPrinter.
type rpcStartDocPrinterRequest struct {
	HPrinter          msrprn.PRINTER_HANDLE
	PDocInfoContainer msrprn.DOC_INFO_CONTAINER
}

func (*rpcStartDocPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcStartDocPrinter }

// rpcStartDocPrinterResponse carries the [out] parameters and return value of RpcStartDocPrinter.
type rpcStartDocPrinterResponse struct {
	PJobId ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// RpcStartDocPrinter calls RpcStartDocPrinter (opnum 17) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcStartDocPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pDocInfoContainer msrprn.DOC_INFO_CONTAINER) (PJobId ndr.DWORD, err error) {
	req := &rpcStartDocPrinterRequest{
		HPrinter:          hPrinter,
		PDocInfoContainer: pDocInfoContainer,
	}
	var resp rpcStartDocPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcStartDocPrinter: %w", err)
		return
	}
	PJobId = resp.PJobId
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcStartDocPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
