package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcIppCreateJobOnPrinterRequest carries the [in] parameters of RpcIppCreateJobOnPrinter.
type rpcIppCreateJobOnPrinterRequest struct {
	HPrinter                    msrprn.PRINTER_HANDLE
	JobId                       ndr.DWORD
	PdlFormat                   *ndr.WSTR `ndr:"unique"`
	JobAttributeGroupBufferSize ndr.DWORD
	JobAttributeGroupBuffer     []uint8 `ndr:"ref,size_is=JobAttributeGroupBufferSize"`
}

func (*rpcIppCreateJobOnPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcIppCreateJobOnPrinter }

// rpcIppCreateJobOnPrinterResponse carries the [out] parameters and return value of RpcIppCreateJobOnPrinter.
type rpcIppCreateJobOnPrinterResponse struct {
	IppResponseBufferSize ndr.DWORD
	IppResponseBuffer     []*uint8  `ndr:"elem=unique,ref,conformant"`
	Status                ndr.DWORD `ndr:"retval"`
}

// RpcIppCreateJobOnPrinter calls RpcIppCreateJobOnPrinter (opnum 119) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcIppCreateJobOnPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, pdlFormat *ndr.WSTR, jobAttributeGroupBufferSize ndr.DWORD, jobAttributeGroupBuffer []uint8) (IppResponseBufferSize ndr.DWORD, IppResponseBuffer []*uint8, err error) {
	req := &rpcIppCreateJobOnPrinterRequest{
		HPrinter:                    hPrinter,
		JobId:                       jobId,
		PdlFormat:                   pdlFormat,
		JobAttributeGroupBufferSize: jobAttributeGroupBufferSize,
		JobAttributeGroupBuffer:     jobAttributeGroupBuffer,
	}
	var resp rpcIppCreateJobOnPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIppCreateJobOnPrinter: %w", err)
		return
	}
	IppResponseBufferSize = resp.IppResponseBufferSize
	IppResponseBuffer = resp.IppResponseBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcIppCreateJobOnPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
