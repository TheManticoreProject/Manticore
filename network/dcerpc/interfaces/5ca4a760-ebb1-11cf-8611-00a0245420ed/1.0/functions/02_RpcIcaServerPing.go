package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcIcaServerPingRequest carries the [in] parameters of RpcIcaServerPing.
type rpcIcaServerPingRequest struct {
	HServer mststs.SERVER_HANDLE
}

func (*rpcIcaServerPingRequest) Opnum() uint16 { return IcaApi.OpnumRpcIcaServerPing }

// rpcIcaServerPingResponse carries the [out] parameters and return value of RpcIcaServerPing.
type rpcIcaServerPingResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcIcaServerPing calls RpcIcaServerPing (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcIcaServerPing(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE) (PResult ndr.DWORD, err error) {
	req := &rpcIcaServerPingRequest{
		HServer: hServer,
	}
	var resp rpcIcaServerPingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIcaServerPing: %w", err)
		return
	}
	PResult = resp.PResult
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcIcaServerPing failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
