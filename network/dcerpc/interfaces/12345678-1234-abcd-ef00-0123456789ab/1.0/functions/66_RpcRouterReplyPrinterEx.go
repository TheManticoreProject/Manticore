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

// rpcRouterReplyPrinterExRequest carries the [in] parameters of RpcRouterReplyPrinterEx.
type rpcRouterReplyPrinterExRequest struct {
	HNotify     msrprn.PRINTER_HANDLE
	DwColor     ndr.DWORD
	FdwFlags    ndr.DWORD
	DwReplyType ndr.DWORD
	Reply       msrprn.RPC_V2_UREPLY_PRINTER
}

func (*rpcRouterReplyPrinterExRequest) Opnum() uint16 { return winspool.OpnumRpcRouterReplyPrinterEx }

// rpcRouterReplyPrinterExResponse carries the [out] parameters and return value of RpcRouterReplyPrinterEx.
type rpcRouterReplyPrinterExResponse struct {
	PdwResult ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcRouterReplyPrinterEx calls RpcRouterReplyPrinterEx (opnum 66) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRouterReplyPrinterEx(rpc ndr.Invoker, hNotify msrprn.PRINTER_HANDLE, dwColor ndr.DWORD, fdwFlags ndr.DWORD, dwReplyType ndr.DWORD, reply msrprn.RPC_V2_UREPLY_PRINTER) (PdwResult ndr.DWORD, err error) {
	req := &rpcRouterReplyPrinterExRequest{
		HNotify:     hNotify,
		DwColor:     dwColor,
		FdwFlags:    fdwFlags,
		DwReplyType: dwReplyType,
		Reply:       reply,
	}
	var resp rpcRouterReplyPrinterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRouterReplyPrinterEx: %w", err)
		return
	}
	PdwResult = resp.PdwResult
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRouterReplyPrinterEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
