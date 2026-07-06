package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetRemoteAddressRequest carries the [in] parameters of RpcGetRemoteAddress.
type rpcGetRemoteAddressRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetRemoteAddressRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetRemoteAddress }

// rpcGetRemoteAddressResponse carries the [out] parameters and return value of RpcGetRemoteAddress.
type rpcGetRemoteAddressResponse struct {
	PRemoteAddress mststs.RCM_REMOTEADDRESS
	Status         ndr.DWORD `ndr:"retval"`
}

// RpcGetRemoteAddress calls RpcGetRemoteAddress (opnum 4) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetRemoteAddress(rpc ndr.Invoker, sessionId ndr.DWORD) (PRemoteAddress mststs.RCM_REMOTEADDRESS, err error) {
	req := &rpcGetRemoteAddressRequest{
		SessionId: sessionId,
	}
	var resp rpcGetRemoteAddressResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetRemoteAddress: %w", err)
		return
	}
	PRemoteAddress = resp.PRemoteAddress
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetRemoteAddress failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
