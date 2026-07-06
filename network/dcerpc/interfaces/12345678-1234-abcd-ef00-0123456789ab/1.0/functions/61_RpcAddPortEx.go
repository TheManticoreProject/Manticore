package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcAddPortExRequest carries the [in] parameters of RpcAddPortEx.
type rpcAddPortExRequest struct {
	PName             *ndr.WSTR `ndr:"unique"`
	PPortContainer    msrprn.PORT_CONTAINER
	PPortVarContainer msrprn.PORT_VAR_CONTAINER
	PMonitorName      ndr.WSTR
}

func (*rpcAddPortExRequest) Opnum() uint16 { return winspool.OpnumRpcAddPortEx }

// rpcAddPortExResponse carries the [out] parameters and return value of RpcAddPortEx.
type rpcAddPortExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPortEx calls RpcAddPortEx (opnum 61) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPortEx(rpc ndr.Invoker, pName *ndr.WSTR, pPortContainer msrprn.PORT_CONTAINER, pPortVarContainer msrprn.PORT_VAR_CONTAINER, pMonitorName ndr.WSTR) (err error) {
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
