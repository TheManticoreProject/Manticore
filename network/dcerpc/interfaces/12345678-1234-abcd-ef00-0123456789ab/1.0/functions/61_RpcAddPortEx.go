package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddPortExRequest carries the [in] parameters of RpcAddPortEx.
type rpcAddPortExRequest struct {
	PName             *ndr.WSTR `ndr:"unique"`
	PPortContainer    structures.PORT_CONTAINER
	PPortVarContainer structures.PORT_VAR_CONTAINER
	PMonitorName      ndr.WSTR
}

func (*rpcAddPortExRequest) Opnum() uint16 { return winspool.OpnumRpcAddPortEx }

// rpcAddPortExResponse carries the [out] parameters and return value of RpcAddPortEx.
type rpcAddPortExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPortEx calls RpcAddPortEx (opnum 61) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPortEx(rpc ndr.Invoker, pName *ndr.WSTR, pPortContainer structures.PORT_CONTAINER, pPortVarContainer structures.PORT_VAR_CONTAINER, pMonitorName ndr.WSTR) (err error) {
	req := &rpcAddPortExRequest{
		PName:             pName,
		PPortContainer:    pPortContainer,
		PPortVarContainer: pPortVarContainer,
		PMonitorName:      pMonitorName,
	}
	var resp rpcAddPortExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPortEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPortEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
