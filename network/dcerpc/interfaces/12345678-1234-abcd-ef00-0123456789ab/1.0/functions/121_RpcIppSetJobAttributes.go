package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcIppSetJobAttributesRequest carries the [in] parameters of RpcIppSetJobAttributes.
type rpcIppSetJobAttributesRequest struct {
	HPrinter                    msrprn.PRINTER_HANDLE
	JobId                       ndr.DWORD
	JobAttributeGroupBufferSize ndr.DWORD
	JobAttributeGroupBuffer     []uint8 `ndr:"ref,size_is=JobAttributeGroupBufferSize"`
}

func (*rpcIppSetJobAttributesRequest) Opnum() uint16 { return winspool.OpnumRpcIppSetJobAttributes }

// rpcIppSetJobAttributesResponse carries the [out] parameters and return value of RpcIppSetJobAttributes.
type rpcIppSetJobAttributesResponse struct {
	IppResponseBufferSize ndr.DWORD
	IppResponseBuffer     []*uint8  `ndr:"elem=unique,ref,conformant"`
	Status                ndr.DWORD `ndr:"retval"`
}

// RpcIppSetJobAttributes calls RpcIppSetJobAttributes (opnum 121) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcIppSetJobAttributes(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, jobAttributeGroupBufferSize ndr.DWORD, jobAttributeGroupBuffer []uint8) (IppResponseBufferSize ndr.DWORD, IppResponseBuffer []*uint8, err error) {
	req := &rpcIppSetJobAttributesRequest{
		HPrinter:                    hPrinter,
		JobId:                       jobId,
		JobAttributeGroupBufferSize: jobAttributeGroupBufferSize,
		JobAttributeGroupBuffer:     jobAttributeGroupBuffer,
	}
	var resp rpcIppSetJobAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIppSetJobAttributes: %w", err)
		return
	}
	IppResponseBufferSize = resp.IppResponseBufferSize
	IppResponseBuffer = resp.IppResponseBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcIppSetJobAttributes failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
