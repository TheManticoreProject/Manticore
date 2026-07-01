package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePortRequest carries the [in] parameters of RpcDeletePort.
type rpcDeletePortRequest struct {
	PName     *ndr.WSTR `ndr:"unique"`
	HWnd      ndr.DWORD
	PPortName ndr.WSTR
}

func (*rpcDeletePortRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePort }

// rpcDeletePortResponse carries the [out] parameters and return value of RpcDeletePort.
type rpcDeletePortResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePort calls RpcDeletePort (opnum 39) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePort(rpc ndr.Invoker, pName *ndr.WSTR, hWnd ndr.DWORD, pPortName ndr.WSTR) (err error) {
	req := &rpcDeletePortRequest{
		PName:     pName,
		HWnd:      hWnd,
		PPortName: pPortName,
	}
	var resp rpcDeletePortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePort: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePort failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
