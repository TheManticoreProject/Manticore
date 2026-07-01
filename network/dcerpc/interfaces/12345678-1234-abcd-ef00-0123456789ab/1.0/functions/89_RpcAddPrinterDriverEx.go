package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddPrinterDriverExRequest carries the [in] parameters of RpcAddPrinterDriverEx.
type rpcAddPrinterDriverExRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PDriverContainer structures.DRIVER_CONTAINER
	DwFileCopyFlags  ndr.DWORD
}

func (*rpcAddPrinterDriverExRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrinterDriverEx }

// rpcAddPrinterDriverExResponse carries the [out] parameters and return value of RpcAddPrinterDriverEx.
type rpcAddPrinterDriverExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPrinterDriverEx calls RpcAddPrinterDriverEx (opnum 89) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrinterDriverEx(rpc ndr.Invoker, pName *ndr.WSTR, pDriverContainer structures.DRIVER_CONTAINER, dwFileCopyFlags ndr.DWORD) (err error) {
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
