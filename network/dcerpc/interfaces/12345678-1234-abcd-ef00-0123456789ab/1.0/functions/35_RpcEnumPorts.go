package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumPortsRequest carries the [in] parameters of RpcEnumPorts.
type rpcEnumPortsRequest struct {
	PName *ndr.WSTR `ndr:"unique"`
	Level ndr.DWORD
	PPort []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf ndr.DWORD
}

func (*rpcEnumPortsRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPorts }

// rpcEnumPortsResponse carries the [out] parameters and return value of RpcEnumPorts.
type rpcEnumPortsResponse struct {
	PPort      []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumPorts calls RpcEnumPorts (opnum 35) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPorts(rpc ndr.Invoker, pName *ndr.WSTR, level ndr.DWORD, pPort []uint8, cbBuf ndr.DWORD) (PPort []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPortsRequest{
		PName: pName,
		Level: level,
		PPort: pPort,
		CbBuf: cbBuf,
	}
	var resp rpcEnumPortsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPorts: %w", err)
		return
	}
	PPort = resp.PPort
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPorts failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
