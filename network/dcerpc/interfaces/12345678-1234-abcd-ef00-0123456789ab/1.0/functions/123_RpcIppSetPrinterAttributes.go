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

// rpcIppSetPrinterAttributesRequest carries the [in] parameters of RpcIppSetPrinterAttributes.
type rpcIppSetPrinterAttributesRequest struct {
	HPrinter                    msrprn.PRINTER_HANDLE
	JobAttributeGroupBufferSize ndr.DWORD
	JobAttributeGroupBuffer     []uint8 `ndr:"ref,size_is=JobAttributeGroupBufferSize"`
}

func (*rpcIppSetPrinterAttributesRequest) Opnum() uint16 {
	return winspool.OpnumRpcIppSetPrinterAttributes
}

// rpcIppSetPrinterAttributesResponse carries the [out] parameters and return value of RpcIppSetPrinterAttributes.
type rpcIppSetPrinterAttributesResponse struct {
	IppResponseBufferSize ndr.DWORD
	IppResponseBuffer     []*uint8  `ndr:"elem=unique,ref,conformant"`
	Status                ndr.DWORD `ndr:"retval"`
}

// RpcIppSetPrinterAttributes calls RpcIppSetPrinterAttributes (opnum 123) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcIppSetPrinterAttributes(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobAttributeGroupBufferSize ndr.DWORD, jobAttributeGroupBuffer []uint8) (IppResponseBufferSize ndr.DWORD, IppResponseBuffer []*uint8, err error) {
	req := &rpcIppSetPrinterAttributesRequest{
		HPrinter:                    hPrinter,
		JobAttributeGroupBufferSize: jobAttributeGroupBufferSize,
		JobAttributeGroupBuffer:     jobAttributeGroupBuffer,
	}
	var resp rpcIppSetPrinterAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIppSetPrinterAttributes: %w", err)
		return
	}
	IppResponseBufferSize = resp.IppResponseBufferSize
	IppResponseBuffer = resp.IppResponseBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcIppSetPrinterAttributes failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
