package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetLoggedOnCountRequest carries the [in] parameters of RpcGetLoggedOnCount.
type rpcGetLoggedOnCountRequest struct {
}

func (*rpcGetLoggedOnCountRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetLoggedOnCount }

// rpcGetLoggedOnCountResponse carries the [out] parameters and return value of RpcGetLoggedOnCount.
type rpcGetLoggedOnCountResponse struct {
	PUserSessions   ndr.DWORD
	PDeviceSessions ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetLoggedOnCount calls RpcGetLoggedOnCount (opnum 15) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetLoggedOnCount(rpc ndr.Invoker) (PUserSessions ndr.DWORD, PDeviceSessions ndr.DWORD, err error) {
	req := &rpcGetLoggedOnCountRequest{}
	var resp rpcGetLoggedOnCountResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetLoggedOnCount: %w", err)
		return
	}
	PUserSessions = resp.PUserSessions
	PDeviceSessions = resp.PDeviceSessions
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetLoggedOnCount failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
