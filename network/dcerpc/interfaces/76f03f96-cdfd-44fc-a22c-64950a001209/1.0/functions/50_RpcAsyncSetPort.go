package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSetPortRequest carries the [in] parameters of RpcAsyncSetPort.
type rpcAsyncSetPortRequest struct {
	PName          *ndr.WSTR `ndr:"unique"`
	PPortName      *ndr.WSTR `ndr:"unique"`
	PPortContainer mspar.PORT_CONTAINER
}

func (*rpcAsyncSetPortRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncSetPort }

// rpcAsyncSetPortResponse carries the [out] parameters and return value of RpcAsyncSetPort.
type rpcAsyncSetPortResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetPort calls RpcAsyncSetPort (opnum 50) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetPort(rpc ndr.Invoker, pName *ndr.WSTR, pPortName *ndr.WSTR, pPortContainer mspar.PORT_CONTAINER) (err error) {
	req := &rpcAsyncSetPortRequest{
		PName:          pName,
		PPortName:      pPortName,
		PPortContainer: pPortContainer,
	}
	var resp rpcAsyncSetPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetPort: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetPort failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
