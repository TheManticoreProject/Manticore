package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetLastInputTimeRequest carries the [in] parameters of RpcGetLastInputTime.
type rpcGetLastInputTimeRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetLastInputTimeRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetLastInputTime }

// rpcGetLastInputTimeResponse carries the [out] parameters and return value of RpcGetLastInputTime.
type rpcGetLastInputTimeResponse struct {
	PLastInputTime int64
	Status         ndr.DWORD `ndr:"retval"`
}

// RpcGetLastInputTime calls RpcGetLastInputTime (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetLastInputTime(rpc ndr.Invoker, sessionId ndr.DWORD) (PLastInputTime int64, err error) {
	req := &rpcGetLastInputTimeRequest{
		SessionId: sessionId,
	}
	var resp rpcGetLastInputTimeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetLastInputTime: %w", err)
		return
	}
	PLastInputTime = resp.PLastInputTime
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetLastInputTime failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
