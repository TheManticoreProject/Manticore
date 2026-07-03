package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcIppGetJobAttributesRequest carries the [in] parameters of RpcIppGetJobAttributes.
type rpcIppGetJobAttributesRequest struct {
	HPrinter           msrprn.PRINTER_HANDLE
	JobId              ndr.DWORD
	AttributeNameCount ndr.DWORD
	AttributeNames     ndr.WSTR
}

func (*rpcIppGetJobAttributesRequest) Opnum() uint16 { return winspool.OpnumRpcIppGetJobAttributes }

// rpcIppGetJobAttributesResponse carries the [out] parameters and return value of RpcIppGetJobAttributes.
type rpcIppGetJobAttributesResponse struct {
	IppResponseBufferSize ndr.DWORD
	IppResponseBuffer     []*uint8  `ndr:"elem=unique,ref,conformant"`
	Status                ndr.DWORD `ndr:"retval"`
}

// RpcIppGetJobAttributes calls RpcIppGetJobAttributes (opnum 120) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcIppGetJobAttributes(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, attributeNameCount ndr.DWORD, attributeNames ndr.WSTR) (IppResponseBufferSize ndr.DWORD, IppResponseBuffer []*uint8, err error) {
	req := &rpcIppGetJobAttributesRequest{
		HPrinter:           hPrinter,
		JobId:              jobId,
		AttributeNameCount: attributeNameCount,
		AttributeNames:     attributeNames,
	}
	var resp rpcIppGetJobAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIppGetJobAttributes: %w", err)
		return
	}
	IppResponseBufferSize = resp.IppResponseBufferSize
	IppResponseBuffer = resp.IppResponseBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcIppGetJobAttributes failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
