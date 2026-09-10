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

// rpcWinStationReInitializeSecurityRequest carries the [in] parameters of RpcWinStationReInitializeSecurity.
type rpcWinStationReInitializeSecurityRequest struct {
	HServer mststs.SERVER_HANDLE
}

func (*rpcWinStationReInitializeSecurityRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationReInitializeSecurity
}

// rpcWinStationReInitializeSecurityResponse carries the [out] parameters and return value of RpcWinStationReInitializeSecurity.
type rpcWinStationReInitializeSecurityResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationReInitializeSecurity calls RpcWinStationReInitializeSecurity (opnum 46) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationReInitializeSecurity(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationReInitializeSecurityRequest{
		HServer: hServer,
	}
	var resp rpcWinStationReInitializeSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationReInitializeSecurity: %w", err)
		return
	}
	PResult = resp.PResult
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcWinStationReInitializeSecurity failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
