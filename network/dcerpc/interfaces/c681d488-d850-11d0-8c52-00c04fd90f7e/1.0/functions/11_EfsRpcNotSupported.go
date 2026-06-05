package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcNotSupportedRequest carries the [in] parameters of EfsRpcNotSupported.
type efsRpcNotSupportedRequest struct {
	Reserved1   *ndr.WSTR `ndr:"unique"`
	Reserved2   *ndr.WSTR `ndr:"unique"`
	DwReserved1 ndr.DWORD
	DwReserved2 ndr.DWORD
	Reserved    *structures.EFS_RPC_BLOB `ndr:"unique"`
	BReserved   ndr.BOOL
}

func (*efsRpcNotSupportedRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcNotSupported }

// efsRpcNotSupportedResponse carries the [out] parameters and return value of EfsRpcNotSupported.
type efsRpcNotSupportedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcNotSupported calls EfsRpcNotSupported (opnum 11) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcNotSupported(rpc ndr.Invoker, reserved1 *ndr.WSTR, reserved2 *ndr.WSTR, dwReserved1 ndr.DWORD, dwReserved2 ndr.DWORD, reserved *structures.EFS_RPC_BLOB, bReserved ndr.BOOL) (err error) {
	req := &efsRpcNotSupportedRequest{
		Reserved1:   reserved1,
		Reserved2:   reserved2,
		DwReserved1: dwReserved1,
		DwReserved2: dwReserved2,
		Reserved:    reserved,
		BReserved:   bReserved,
	}
	var resp efsRpcNotSupportedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcNotSupported: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcNotSupported failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
