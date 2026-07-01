package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcIppGetPrinterAttributesRequest carries the [in] parameters of RpcIppGetPrinterAttributes.
type rpcIppGetPrinterAttributesRequest struct {
	HPrinter           structures.PRINTER_HANDLE
	AttributeNameCount ndr.DWORD
	AttributeNames     ndr.WSTR
}

func (*rpcIppGetPrinterAttributesRequest) Opnum() uint16 {
	return winspool.OpnumRpcIppGetPrinterAttributes
}

// rpcIppGetPrinterAttributesResponse carries the [out] parameters and return value of RpcIppGetPrinterAttributes.
type rpcIppGetPrinterAttributesResponse struct {
	IppResponseBufferSize ndr.DWORD
	IppResponseBuffer     []*uint8  `ndr:"elem=unique,ref,conformant"`
	Status                ndr.DWORD `ndr:"retval"`
}

// RpcIppGetPrinterAttributes calls RpcIppGetPrinterAttributes (opnum 122) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcIppGetPrinterAttributes(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, attributeNameCount ndr.DWORD, attributeNames ndr.WSTR) (IppResponseBufferSize ndr.DWORD, IppResponseBuffer []*uint8, err error) {
	req := &rpcIppGetPrinterAttributesRequest{
		HPrinter:           hPrinter,
		AttributeNameCount: attributeNameCount,
		AttributeNames:     attributeNames,
	}
	var resp rpcIppGetPrinterAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIppGetPrinterAttributes: %w", err)
		return
	}
	IppResponseBufferSize = resp.IppResponseBufferSize
	IppResponseBuffer = resp.IppResponseBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcIppGetPrinterAttributes failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
