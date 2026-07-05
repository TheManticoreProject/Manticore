package functions

import (
	"fmt"

	SessEnvPublicRpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1257b580-ce2f-4109-82d6-a9459d0bf6bc/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcShadow2Request carries the [in] parameters of RpcShadow2.
type rpcShadow2Request struct {
	TargetSessionId    ndr.DWORD
	ERequestControl    mststs.SHADOW_CONTROL_REQUEST
	ERequestPermission mststs.SHADOW_PERMISSION_REQUEST
	CchInvitation      ndr.DWORD
}

func (*rpcShadow2Request) Opnum() uint16 { return SessEnvPublicRpc.OpnumRpcShadow2 }

// rpcShadow2Response carries the [out] parameters and return value of RpcShadow2.
type rpcShadow2Response struct {
	PePermission  mststs.SHADOW_REQUEST_RESPONSE
	PszInvitation ndr.WSTR
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcShadow2 calls RpcShadow2 (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcShadow2(rpc ndr.Invoker, targetSessionId ndr.DWORD, eRequestControl mststs.SHADOW_CONTROL_REQUEST, eRequestPermission mststs.SHADOW_PERMISSION_REQUEST, cchInvitation ndr.DWORD) (PePermission mststs.SHADOW_REQUEST_RESPONSE, PszInvitation ndr.WSTR, err error) {
	req := &rpcShadow2Request{
		TargetSessionId:    targetSessionId,
		ERequestControl:    eRequestControl,
		ERequestPermission: eRequestPermission,
		CchInvitation:      cchInvitation,
	}
	var resp rpcShadow2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcShadow2: %w", err)
		return
	}
	PePermission = resp.PePermission
	PszInvitation = resp.PszInvitation
	if uint32(resp.Status) != SessEnvPublicRpc.StatusSuccess {
		err = fmt.Errorf("RpcShadow2 failed: %s", SessEnvPublicRpc.StatusString(uint32(resp.Status)))
	}
	return
}
