package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPortsRequest carries the [in] parameters of RpcAsyncEnumPorts.
type rpcAsyncEnumPortsRequest struct {
	PName *ndr.WSTR `ndr:"unique"`
	Level ndr.DWORD
	PPort []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf ndr.DWORD
}

func (*rpcAsyncEnumPortsRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncEnumPorts }

// rpcAsyncEnumPortsResponse carries the [out] parameters and return value of RpcAsyncEnumPorts.
type rpcAsyncEnumPortsResponse struct {
	PPort      []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPorts calls RpcAsyncEnumPorts (opnum 47) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPorts(rpc ndr.Invoker, pName *ndr.WSTR, level ndr.DWORD, pPort []uint8, cbBuf ndr.DWORD) (PPort []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPortsRequest{
		PName: pName,
		Level: level,
		PPort: pPort,
		CbBuf: cbBuf,
	}
	var resp rpcAsyncEnumPortsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPorts: %w", err)
		return
	}
	PPort = resp.PPort
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPorts failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
