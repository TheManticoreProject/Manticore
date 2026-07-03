package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcSetPortRequest carries the [in] parameters of RpcSetPort.
type rpcSetPortRequest struct {
	PName          *ndr.WSTR `ndr:"unique"`
	PPortName      *ndr.WSTR `ndr:"unique"`
	PPortContainer msrprn.PORT_CONTAINER
}

func (*rpcSetPortRequest) Opnum() uint16 { return winspool.OpnumRpcSetPort }

// rpcSetPortResponse carries the [out] parameters and return value of RpcSetPort.
type rpcSetPortResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetPort calls RpcSetPort (opnum 71) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetPort(rpc ndr.Invoker, pName *ndr.WSTR, pPortName *ndr.WSTR, pPortContainer msrprn.PORT_CONTAINER) (err error) {
	req := &rpcSetPortRequest{
		PName:          pName,
		PPortName:      pPortName,
		PPortContainer: pPortContainer,
	}
	var resp rpcSetPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetPort: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetPort failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
