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

// rpcGetConfigDataRequest carries the [in] parameters of RpcGetConfigData.
type rpcGetConfigDataRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetConfigDataRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetConfigData }

// rpcGetConfigDataResponse carries the [out] parameters and return value of RpcGetConfigData.
type rpcGetConfigDataResponse struct {
	PpBuff          []uint8 `ndr:"unique,conformant"`
	POutBuffByteLen ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetConfigData calls RpcGetConfigData (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetConfigData(rpc ndr.Invoker, sessionId ndr.DWORD) (PpBuff []uint8, POutBuffByteLen ndr.DWORD, err error) {
	req := &rpcGetConfigDataRequest{
		SessionId: sessionId,
	}
	var resp rpcGetConfigDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetConfigData: %w", err)
		return
	}
	PpBuff = resp.PpBuff
	POutBuffByteLen = resp.POutBuffByteLen
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetConfigData failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
