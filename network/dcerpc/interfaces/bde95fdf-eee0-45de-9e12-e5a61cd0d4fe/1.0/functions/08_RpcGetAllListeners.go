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

// rpcGetAllListenersRequest carries the [in] parameters of RpcGetAllListeners.
type rpcGetAllListenersRequest struct {
	Level ndr.DWORD
}

func (*rpcGetAllListenersRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetAllListeners }

// rpcGetAllListenersResponse carries the [out] parameters and return value of RpcGetAllListeners.
type rpcGetAllListenersResponse struct {
	PpListeners   []mststs.LISTENERENUM `ndr:"unique,conformant"`
	PNumListeners ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcGetAllListeners calls RpcGetAllListeners (opnum 8) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetAllListeners(rpc ndr.Invoker, level ndr.DWORD) (PpListeners []mststs.LISTENERENUM, PNumListeners ndr.DWORD, err error) {
	req := &rpcGetAllListenersRequest{
		Level: level,
	}
	var resp rpcGetAllListenersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetAllListeners: %w", err)
		return
	}
	PpListeners = resp.PpListeners
	PNumListeners = resp.PNumListeners
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetAllListeners failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
